<#
.SYNOPSIS
    Loads the Calpeda Specs workbook into Lightspeed in one run.

.DESCRIPTION
    1. Reads and checks every sheet of the workbook (generate_calpeda_specs.py).
    2. Writes the specs to the database named in the API's .env file, in one transaction:
       voltage/FLA, impeller, the 9 PUMP spec rows, connection type and the special-item flag.
       Nothing is written unless every workbook model matches exactly one Calpeda item.
    3. Clears that database's API cache in Redis - the same keys GET /api/admin/clear_all
       clears - so Item Entry shows the new specs straight away instead of within the hour.

    Without -Commit it is a trial run: everything is done and counted, then rolled back, and
    the cache is left alone. Safe to run again whenever the workbook changes. Takes a minute or
    two for the full workbook.

    -AddMissingItems is for a database that never had the Calpeda item load: in the same
    transaction it first adds every pump with no item yet (from -ItemsExcel, the Data Entry
    workbook) and every impeller code missing from Material setup. Nothing that already exists
    is changed or removed.

.EXAMPLE
    .\load_calpeda_specs.ps1
    Trial run against the database in D:\ERP_API\.env (the 3000 API). Nothing is kept.

.EXAMPLE
    .\load_calpeda_specs.ps1 -Commit
    Loads it for real, then clears the cache.

.EXAMPLE
    .\load_calpeda_specs.ps1 -Excel 'D:\new\Calpeda Specs.xlsx' -EnvFile 'D:\Gideon\ERP_API\.env' -Commit
    A different workbook, into the database of a different API instance.

.EXAMPLE
    .\load_calpeda_specs.ps1 -EnvFile 'D:\Gideon\ERP_API\.env' -AddMissingItems -Commit
    The 3001 API's database: adds the missing Calpeda pumps and impeller codes, then the specs.

.EXAMPLE
    .\load_calpeda_specs.ps1 -CacheOnly
    Only clears the API cache for the database in -EnvFile.
#>
param(
    [string]$Excel = 'D:\Gideon\data migration excel format\Calpeda Specs (3_17_26).xlsx',
    # The .env of the API instance whose database gets the specs. D:\ERP_API is the 3000 API.
    [string]$EnvFile = 'D:\ERP_API\.env',
    [switch]$Commit,
    [switch]$AddMissingItems,
    [string]$ItemsExcel = 'D:\Gideon\data migration excel format\Calpeda Data Entry (3_17_26).xlsx',
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

# --- A minimal Redis client: enough for SCAN and DEL, so no redis-cli is needed. ---

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

# Removes every cached query of one database. The API keys them db:<DB_HOST>/<DB_NAME>:model:...
# (services/cache_namespace.go), and clear_all removes model:* inside that namespace; glob
# characters in the namespace are escaped the same way, so no other database's cache is touched.
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
if ($AddMissingItems -and -not (Test-Path -LiteralPath $ItemsExcel)) { throw "Data Entry workbook not found: $ItemsExcel" }

if ($Commit) { Write-Host "LOADING Calpeda specs into $target" }
else { Write-Host "TRIAL RUN of the Calpeda specs load against $target - nothing will be kept" }

$sqlFile = Join-Path ([IO.Path]::GetTempPath()) ('load_calpeda_specs_{0}.sql' -f [guid]::NewGuid().ToString('N'))
$text = ''
try {
    # Native tools report through their exit code; don't let PowerShell turn their stderr into
    # a terminating error first.
    $ErrorActionPreference = 'Continue'

    # 1. Read and check the workbook, and build the load.
    $genArgs = @((Join-Path $PSScriptRoot 'generate_calpeda_specs.py'), $Excel, $sqlFile)
    if ($Commit) { $genArgs += '--commit' }
    if ($AddMissingItems) { $genArgs += @('--add-missing-items', $ItemsExcel) }
    & python @genArgs
    if ($LASTEXITCODE -ne 0) { throw 'The workbook failed its checks (see above) - nothing was changed.' }

    # 2. Run it. -b makes sqlcmd stop, and fail, on the load's own RAISERRORs.
    $server = "$($envValues['DB_HOST']),$($envValues['DB_PORT'])"
    $output = & sqlcmd -S $server -d $envValues['DB_NAME'] -U $envValues['DB_USERNAME'] -P $envValues['DB_PASSWORD'] -i $sqlFile -b -W -s '|' -f 65001 2>&1
    $exit = $LASTEXITCODE
    $text = ($output | Out-String).Replace($envValues['DB_PASSWORD'], '****')
    Write-Host $text
    if ($exit -ne 0) { throw 'The database load stopped (see above) - nothing was changed.' }
}
finally {
    $ErrorActionPreference = 'Stop'
    if (Test-Path -LiteralPath $sqlFile) { Remove-Item -LiteralPath $sqlFile -Force }
}

if (-not $Commit) {
    Write-Host 'Trial run only - nothing was kept. Run again with -Commit to load it.'
    return
}
if ($text -notmatch 'COMMITTED') { throw 'The load did not report COMMITTED - check the messages above.' }

# 3. Clear the cache, so Item Entry reads the new specs now.
try {
    $count = Clear-ApiCache $envValues['DB_HOST'] $envValues['DB_NAME']
    Write-Host "Cleared $count cached API entries for $target. Item Entry shows the new specs now."
}
catch {
    Write-Warning "The specs are loaded, but the API cache could not be cleared: $($_.Exception.Message). Run this script with -CacheOnly, restart the API, or wait up to an hour."
}
