<#
.SYNOPSIS
    Replaces the whole chart of accounts with the '142 training' tab of
    'GL ACCOUNTS - CLASSIFIED.xlsx'.

.DESCRIPTION
    1. Reads and checks the tab (generate_chart_of_accounts.py).
    2. In one transaction on the database named in -EnvFile:
         - adds a Chart Class for any of ASSET / LIABILITY / EQUITY / REVENUE / EXPENSE missing
         - DELETES every row of tbl_setup_chart_of_accounts
         - inserts the tab's 226 accounts, sub-accounts grouped under their account
         - keeps the seven row ids the posting code looks up directly:
             40030 ACCOUNTS PAYABLE   -> 2000001-1     70034 CASH ON HAND    -> 1000001-5
             50030 NON-TRADE EXPENSE  -> 5000021-2     70035 CASH ON BANK    -> CASH
             70032 TRADE RECEIVABLE   -> 1000002-1     70037 SALES           -> 4000001-5
             70038 ADVANCE PAYMENT    -> 2000008-2
         - logs a DELETE and an INSERT row per account in z_tbl_setup_chart_of_accounts_at
         - lists tax setup, BPI, bulk IR and journal rows left pointing at a deleted account
           (reported only, never changed)
       Nothing is kept unless every check passes.
    3. Clears that database's API cache in Redis, so Accounting shows the new chart at once.

    Without -Commit it is a trial run: everything is done and counted, then rolled back, and
    the cache is left alone.

.EXAMPLE
    .\load_chart_of_accounts.ps1 -EnvFile 'D:\Gideon\ERP_API\.env'
    Trial run against the 3001 API's database. Nothing is kept.

.EXAMPLE
    .\load_chart_of_accounts.ps1 -EnvFile 'D:\ERP_API\.env' -Commit
    Replaces the chart in the 3000 API's database, then clears its cache.

.EXAMPLE
    .\load_chart_of_accounts.ps1 -EnvFile 'D:\ERP_API\.env' -CacheOnly
    Only clears the API cache for that database.
#>
param(
    # The .env of the API instance whose database gets the chart - there is deliberately no
    # default, because this deletes the chart that is there. D:\ERP_API is the 3000 API,
    # D:\Gideon\ERP_API the 3001 API.
    [Parameter(Mandatory = $true)]
    [string]$EnvFile,
    [string]$Excel = 'D:\Gideon\GL ACCOUNTS - CLASSIFIED.xlsx',
    [switch]$Commit,
    [switch]$CacheOnly,
    [string]$RedisHost = 'localhost',
    [int]$RedisPort = 6379
)

$ErrorActionPreference = 'Stop'

function Read-EnvFile([string]$path, [string[]]$required) {
    if (-not (Test-Path -LiteralPath $path)) { throw "No .env file at $path" }
    $values = @{}
    foreach ($line in Get-Content -LiteralPath $path) {
        if ($line -match '^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*?)\s*$') {
            $values[$matches[1]] = $matches[2].Trim('"', "'")
        }
    }
    foreach ($key in $required) {
        if (-not $values[$key]) { throw "$path has no $key" }
    }
    return $values
}

# --- A minimal Redis client (same as load_calpeda_specs.ps1): enough for SCAN and DEL. ---

function Read-RedisLine($stream) {
    $buffer = New-Object System.IO.MemoryStream
    while ($true) {
        $b = $stream.ReadByte()
        if ($b -lt 0) { throw 'Redis closed the connection' }
        if ($b -eq 13) { [void]$stream.ReadByte(); break }
        $buffer.WriteByte([byte]$b)
    }
    return [Text.Encoding]::UTF8.GetString($buffer.ToArray())
}

function Read-RedisReply($stream) {
    $line = Read-RedisLine $stream
    $rest = $line.Substring(1)
    switch ($line.Substring(0, 1)) {
        '+' { return $rest }
        '-' { throw "Redis error: $rest" }
        ':' { return [long]$rest }
        '$' {
            $length = [int]$rest
            if ($length -lt 0) { return $null }
            $data = New-Object byte[] $length
            $read = 0
            while ($read -lt $length) {
                $n = $stream.Read($data, $read, $length - $read)
                if ($n -le 0) { throw 'Redis closed the connection' }
                $read += $n
            }
            [void]$stream.ReadByte(); [void]$stream.ReadByte()
            return [Text.Encoding]::UTF8.GetString($data)
        }
        '*' {
            $count = [int]$rest
            $items = New-Object System.Collections.ArrayList
            for ($i = 0; $i -lt $count; $i++) { [void]$items.Add((Read-RedisReply $stream)) }
            return ,$items.ToArray()
        }
        default { throw "Unexpected Redis reply: $line" }
    }
}

function Send-RedisCommand($stream, [string[]]$parts) {
    $sb = New-Object System.Text.StringBuilder
    [void]$sb.Append('*').Append($parts.Count).Append("`r`n")
    foreach ($part in $parts) {
        [void]$sb.Append('$').Append([Text.Encoding]::UTF8.GetByteCount($part)).Append("`r`n").Append($part).Append("`r`n")
    }
    $bytes = [Text.Encoding]::UTF8.GetBytes($sb.ToString())
    $stream.Write($bytes, 0, $bytes.Length)
    $stream.Flush()
    return Read-RedisReply $stream
}

# Removes every cached query of one database (keys db:<DB_HOST>/<DB_NAME>:model:..., the same
# ones GET /api/admin/clear_all removes). No other database's cache is touched.
function Clear-ApiCache([string]$dbHost, [string]$dbName) {
    $namespace = "db:${dbHost}/${dbName}:"
    $escaped = -join ($namespace.ToCharArray() | ForEach-Object { if ('*?[]\'.Contains([string]$_)) { '\' + $_ } else { [string]$_ } })
    $pattern = $escaped + 'model:*'

    $client = New-Object System.Net.Sockets.TcpClient
    try {
        $client.Connect($RedisHost, $RedisPort)
        $stream = New-Object System.IO.BufferedStream($client.GetStream())
        $cursor = '0'
        $deleted = 0
        do {
            $reply = Send-RedisCommand $stream @('SCAN', $cursor, 'MATCH', $pattern, 'COUNT', '1000')
            $cursor = [string]$reply[0]
            $keys = @($reply[1])
            if ($keys.Count -gt 0) {
                $deleted += [long](Send-RedisCommand $stream (@('DEL') + $keys))
            }
        } while ($cursor -ne '0')
        return $deleted
    }
    finally {
        $client.Close()
    }
}

# --- Main ---

$required = if ($CacheOnly) { @('DB_HOST', 'DB_NAME') } else { @('DB_HOST', 'DB_PORT', 'DB_NAME', 'DB_USERNAME', 'DB_PASSWORD') }
$envValues = Read-EnvFile $EnvFile $required
$target = "$($envValues['DB_NAME']) on $($envValues['DB_HOST'])"

if ($CacheOnly) {
    $count = Clear-ApiCache $envValues['DB_HOST'] $envValues['DB_NAME']
    Write-Host "Cleared $count cached API entries for $target."
    return
}

foreach ($tool in 'python', 'sqlcmd') {
    if (-not (Get-Command $tool -ErrorAction SilentlyContinue)) { throw "$tool is not installed or not on the PATH." }
}
if (-not (Test-Path -LiteralPath $Excel)) { throw "Workbook not found: $Excel" }

if ($Commit) { Write-Host "REPLACING the chart of accounts in $target" }
else { Write-Host "TRIAL RUN of the chart of accounts replacement against $target - nothing will be kept" }

$sqlFile = Join-Path ([IO.Path]::GetTempPath()) ('load_chart_of_accounts_{0}.sql' -f [guid]::NewGuid().ToString('N'))
$text = ''
try {
    # Native tools report through their exit code; don't let PowerShell turn their stderr into
    # a terminating error first.
    $ErrorActionPreference = 'Continue'

    # 1. Read and check the tab, and build the load.
    $genArgs = @((Join-Path $PSScriptRoot 'generate_chart_of_accounts.py'), $Excel, $sqlFile)
    if ($Commit) { $genArgs += '--commit' }
    & python @genArgs
    if ($LASTEXITCODE -ne 0) { throw 'The tab failed its checks (see above) - nothing was changed.' }

    # 2. Run it. -b makes sqlcmd stop, and fail, on the load's own RAISERRORs.
    $server = "$($envValues['DB_HOST']),$($envValues['DB_PORT'])"
    $output = & sqlcmd -S $server -d $envValues['DB_NAME'] -U $envValues['DB_USERNAME'] -P $envValues['DB_PASSWORD'] -i $sqlFile -b -W -s '|' -f 65001 2>&1
    $exit = $LASTEXITCODE
    $text = ($output | Out-String).Replace($envValues['DB_PASSWORD'], '****')
    Write-Host $text
    if ($exit -ne 0) { throw 'The load stopped (see above) - nothing was changed.' }
}
finally {
    $ErrorActionPreference = 'Stop'
    if (Test-Path -LiteralPath $sqlFile) { Remove-Item -LiteralPath $sqlFile -Force }
}

if (-not $Commit) {
    Write-Host 'Trial run only - nothing was kept. Run again with -Commit to replace the chart.'
    return
}
if ($text -notmatch 'COMMITTED') { throw 'The load did not report COMMITTED - check the messages above.' }

# 3. Clear the cache, so the new chart shows now.
try {
    $count = Clear-ApiCache $envValues['DB_HOST'] $envValues['DB_NAME']
    Write-Host "Cleared $count cached API entries for $target. Accounting shows the new chart now."
}
catch {
    Write-Warning "The chart is replaced, but the API cache could not be cleared: $($_.Exception.Message). Run this script with -CacheOnly, restart the API, or wait up to an hour."
}
