# ---------------------------------------------------------------------------
# start-tunnel-server.ps1 - Lightspeed ERP test tunnel, SERVER edition
#
# Starts a Cloudflare "quick tunnel" in front of the local ERP API.
#
# Unlike the workstation version this does NOT rewrite any client config: the
# server runs the API only, and the desktop apps live elsewhere. You take the
# printed URL and put it into the clients yourself.
#
# This is a TESTING aid. A server should end up on a NAMED tunnel installed as a
# Windows service - fixed hostname, survives reboot, no URL to copy each time:
#   cloudflared.exe service install <token from the Zero Trust dashboard>
# ---------------------------------------------------------------------------

$ErrorActionPreference = 'Stop'

# Windows Server 2012 / .NET Framework defaults ServicePointManager to SSL3 and
# TLS 1.0. Cloudflare requires TLS 1.2 or better, so every Invoke-WebRequest to
# the tunnel fails at the handshake - and because that throws with no
# .Exception.Response, it is easily mistaken for a DNS failure. Opt in explicitly
# rather than relying on the OS default, which varies by build and patch level.
try {
    [Net.ServicePointManager]::SecurityProtocol =
        [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
} catch {
    Write-Host '  [warn] Could not enable TLS 1.2 - this PowerShell may be too old.' -ForegroundColor Yellow
}

$Root           = 'C:\Users\Administrator\Desktop\ERPInstaller\cloudflared'
$CloudflaredExe = Join-Path $Root 'cloudflared.exe'
$LogFile        = Join-Path $Root 'tunnel.log'
$LocalApi       = 'http://localhost:3000'

function Say($msg, $colour) {
    if ($colour) { Write-Host $msg -ForegroundColor $colour } else { Write-Host $msg }
}

# The interface actually carrying internet traffic - the one with a default
# route, lowest metric first. Hardcoding "Wi-Fi" is wrong on a server, which
# usually has Ethernet and may have several NICs.
function Get-InternetInterface {
    try {
        $r = Get-NetRoute -DestinationPrefix '0.0.0.0/0' -ErrorAction Stop |
             Sort-Object RouteMetric | Select-Object -First 1
        if ($r) { return $r.InterfaceAlias }
    } catch { }
    return $null
}

Say ''
Say '=== Lightspeed ERP test tunnel (server) ===' 'Cyan'
Say ''

# --- 1. Sanity checks ------------------------------------------------------
if (-not (Test-Path $CloudflaredExe)) {
    Say "cloudflared.exe not found at $CloudflaredExe" 'Red'
    exit 1
}

# Is the API actually up? A tunnel to a dead port returns 502 for everything,
# which looks exactly like a broken tunnel from inside the app.
$apiUp = $false
try {
    Invoke-WebRequest -Uri $LocalApi -TimeoutSec 5 -UseBasicParsing | Out-Null
    $apiUp = $true
} catch {
    # Fiber answers 404 on "/" because no route is registered there. That is a
    # LIVE server, so treat any HTTP response as success and only a transport
    # failure as down.
    if ($_.Exception.Response) { $apiUp = $true }
}
if ($apiUp) {
    Say "  [ok]   API is responding on $LocalApi" 'Green'
} else {
    Say "  [warn] Nothing is answering on $LocalApi" 'Yellow'
    Say "         Start the API first, or every request through the tunnel returns 502." 'Yellow'
}

# --- 2. Stop any tunnel already running ------------------------------------
$existing = Get-Process -Name 'cloudflared' -ErrorAction SilentlyContinue
if ($existing) {
    Say "  [..]   Stopping $($existing.Count) running tunnel(s)"
    $existing | Stop-Process -Force
    Start-Sleep -Seconds 2
}

# --- 3. Start the tunnel ---------------------------------------------------
if (Test-Path $LogFile) { Remove-Item $LogFile -Force }

Say "  [..]   Starting tunnel to $LocalApi"
Start-Process -FilePath $CloudflaredExe `
    -ArgumentList @('tunnel', '--url', $LocalApi, '--logfile', $LogFile, '--loglevel', 'info') `
    -WindowStyle Hidden

# --- 4. Wait for Cloudflare to assign a hostname ---------------------------
$tunnelUrl = $null
for ($i = 0; $i -lt 45; $i++) {
    Start-Sleep -Seconds 1
    if (-not (Test-Path $LogFile)) { continue }
    $m = Select-String -Path $LogFile -Pattern 'https://[a-z0-9-]+\.trycloudflare\.com' -ErrorAction SilentlyContinue
    if ($m) {
        $tunnelUrl = $m[-1].Matches[0].Value
        break
    }
}

if (-not $tunnelUrl) {
    Say ''
    Say '  [FAIL] Cloudflare never assigned a URL.' 'Red'
    Say "         Check $LogFile for the reason." 'Red'
    exit 1
}

$tunnelHost = ([Uri]$tunnelUrl).Host
Say "  [ok]   Tunnel is up" 'Green'
Say ''
Say "         $tunnelUrl" 'Cyan'
Say ''
Say '         Put that URL into the client configs as:' 'Gray'
Say "           ApiBaseUrl.Production = $tunnelUrl/api" 'Gray'
Say "           WssBaseUrl.Production = $($tunnelUrl -replace '^https','wss')/api/ws" 'Gray'
Say ''

# --- 5. Prove it end to end ------------------------------------------------
# Retries, because a freshly-minted hostname takes a few seconds to appear in
# whichever resolver this machine is pointed at.
$verified = $false
for ($try = 1; $try -le 6; $try++) {
    try {
        $r = Invoke-WebRequest -Uri "$tunnelUrl/api/login" -Method POST `
                -ContentType 'application/json' -Body '{}' -TimeoutSec 20 -UseBasicParsing
        Say "  [ok]   API reachable through tunnel (HTTP $($r.StatusCode))" 'Green'
        $verified = $true
        break
    } catch {
        if ($_.Exception.Response) {
            $code = [int]$_.Exception.Response.StatusCode
            if ($code -eq 401) {
                Say "  [ok]   API reachable through tunnel (HTTP 401 - login rejected an empty body, as expected)" 'Green'
                $verified = $true
            } elseif ($code -eq 502) {
                Say "  [FAIL] Tunnel is up but the API is not answering (HTTP 502). Start the API." 'Red'
                $verified = $true   # answered; not a DNS problem
            } else {
                Say "  [warn] Tunnel returned HTTP $code" 'Yellow'
                $verified = $true
            }
            break
        }
        $lastError = $_.Exception.Message
        if ($try -lt 6) { Start-Sleep -Seconds 5 }
    }
}

if (-not $verified) {
    Say ''
    Say "  [FAIL] Could not reach $tunnelHost from this machine." 'Red'

    # Before blaming the network, check the tunnel is still running. A quick
    # tunnel's hostname exists only while its cloudflared process does - when
    # that exits, Cloudflare deregisters the name and it stops resolving
    # entirely (HostNotFound), which reads exactly like a DNS fault but is not.
    $stillUp = Get-Process -Name 'cloudflared' -ErrorAction SilentlyContinue
    if (-not $stillUp) {
        Say '' 
        Say '  [!!]   cloudflared is NOT RUNNING any more.' 'Red'
        Say '         The tunnel started, got a hostname, then exited - so the name' 'Yellow'
        Say '         was deregistered and nothing can resolve it. This is not DNS.' 'Yellow'
        Say ''
        Say "         Look at the end of $LogFile for why it quit:" 'Cyan'
        Say ("           Get-Content '" + $LogFile + "' -Tail 30") 'Gray'
        Say ''
        Say '         A quick tunnel also dies with the machine and takes its' 'Yellow'
        Say '         hostname with it. For a server, install a NAMED tunnel as a' 'Yellow'
        Say '         service instead - fixed hostname, restarts on boot:' 'Yellow'
        Say '           cloudflared.exe service install <token>' 'Gray'
        Say ''
        return
    }

    Say '         cloudflared is still running - this is the connection out.' 'Yellow'
    if ($lastError) {
        Say ''
        Say "         Error: $lastError" 'Red'
    }
    Say ''

    $iface = Get-InternetInterface
    $dns   = $null
    if ($iface) {
        try {
            $dns = (Get-DnsClientServerAddress -InterfaceAlias $iface -AddressFamily IPv4 -ErrorAction Stop).ServerAddresses
        } catch { }
    }

    $v4 = $null; $v6 = $null; $pub = $null
    try { $v4  = Resolve-DnsName $tunnelHost -Type A    -ErrorAction Stop } catch { }
    try { $v6  = Resolve-DnsName $tunnelHost -Type AAAA -ErrorAction Stop } catch { }
    try { $pub = Resolve-DnsName $tunnelHost -Type A -Server 1.1.1.1 -ErrorAction Stop } catch { }

    Say '         Diagnosis:' 'Yellow'
    Say ("           interface       : " + $(if ($iface) { $iface } else { '(could not determine)' })) 'Yellow'
    Say ("           its DNS servers : " + $(if ($dns) { $dns -join ', ' } else { '(none reported)' })) 'Yellow'
    Say ("           A record here   : " + $(if ($v4)  { ($v4  | Where-Object IPAddress | ForEach-Object IPAddress) -join ', ' } else { 'NOT RETURNED' })) 'Yellow'
    Say ("           AAAA record here: " + $(if ($v6)  { ($v6  | Where-Object IPAddress | ForEach-Object IPAddress) -join ', ' } else { 'not returned' })) 'Yellow'
    Say ("           A via 1.1.1.1   : " + $(if ($pub) { ($pub | Where-Object IPAddress | ForEach-Object IPAddress) -join ', ' } else { 'NOT RETURNED' })) 'Yellow'
    Say ''

    if ($pub -and -not $v4) {
        # Cloudflare can resolve it and the local resolver cannot - the local
        # resolver is the fault, whether it returns nothing or IPv6 only.
        Say '         Cause: this machine''s DNS server cannot resolve the name, but' 'Cyan'
        Say '         Cloudflare''s own DNS can. The local resolver is the problem.' 'Cyan'
        Say ''
        if ($iface) {
            Say '         Fix (run as Administrator):' 'Cyan'
            Say ("           Set-DnsClientServerAddress -InterfaceAlias `"$iface`" -ServerAddresses 1.1.1.1,8.8.8.8") 'Cyan'
            Say  '           Clear-DnsClientCache' 'Cyan'
            Say ''
            Say '         To undo, hand DNS back to the network:' 'Gray'
            Say ("           Set-DnsClientServerAddress -InterfaceAlias `"$iface`" -ResetServerAddresses") 'Gray'
            Say ''
            Say '         NOTE: if this server resolves other machines by name (a' 'Gray'
            Say '         domain controller, a file share), changing DNS may break that.' 'Gray'
            Say '         Adding 1.1.1.1 as a SECOND entry after the existing server is' 'Gray'
            Say '         the safer move on a domain-joined box.' 'Gray'
        }
    } elseif (-not $pub -and -not $v4) {
        Say '         Cause: not resolvable even via 1.1.1.1 - so either outbound DNS' 'Cyan'
        Say '         (UDP 53) is blocked from this machine, or the hostname genuinely' 'Cyan'
        Say '         has not propagated yet. Wait a minute and run this again.' 'Cyan'
    } else {
        # Name resolution is working, so the failure is further along: the
        # handshake or the socket itself.
        Say '         DNS is FINE - the name resolves here and matches 1.1.1.1.' 'Cyan'
        Say '         So this is the connection, not the lookup. In order of likelihood:' 'Cyan'
        Say ''
        Say '         1. TLS. Windows Server 2012 defaults .NET to TLS 1.0, which' 'Cyan'
        Say '            Cloudflare refuses. This script now opts in to TLS 1.2 for' 'Cyan'
        Say '            itself, but the OS-wide default needs a registry change:' 'Cyan'
        Say '              https://learn.microsoft.com/security/engineering/solving-tls1-problem' 'Gray'
        Say '            Quick check - this should print True:' 'Cyan'
        Say '              [Net.ServicePointManager]::SecurityProtocol -band [Net.SecurityProtocolType]::Tls12' 'Gray'
        Say ''
        Say '         2. Outbound HTTPS blocked. Test the socket directly:' 'Cyan'
        Say ("              Test-NetConnection " + $tunnelHost + " -Port 443") 'Gray'
        Say '            TcpTestSucceeded False means a firewall or proxy is in the way.' 'Cyan'
        Say ''
        Say '         3. A proxy that PowerShell is not using. If this network needs' 'Cyan'
        Say '            one, Invoke-WebRequest needs -Proxy to be told about it.' 'Cyan'
        Say ''
        Say '         NOTE: the tunnel itself is UP regardless - cloudflared makes its' 'Yellow'
        Say '         own outbound QUIC/HTTPS connection and already succeeded. Clients' 'Yellow'
        Say '         elsewhere may well reach the URL fine even if this check cannot.' 'Yellow'
    }
}

Say ''
Say '  The tunnel runs in the background. It stays up until you reboot,' 'Gray'
Say '  sleep the machine, or run stop-tunnel.bat. Closing this window is fine.' 'Gray'
Say ''
