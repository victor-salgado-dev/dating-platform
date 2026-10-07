[CmdletBinding()]
param(
    [switch]$PrepareOnlineUsers,
    [string]$BaseUrl = 'http://localhost:8080',
    [int]$Users = 100,
    [int]$MaxPage = 3,
    [switch]$SkipExactTotal,
    [ValidateSet('random', 'stable')]
    [string]$DiscoverFilterMode = 'random',
    [int]$SampleIntervalSeconds = 3,
    [string]$PostgresContainer = 'dating-platform-postgres-1',
    [string]$BackendContainer = 'dating-platform-backend-1'
)

$ErrorActionPreference = 'Stop'
$stamp = Get-Date -Format 'yyyyMMdd_HHmmss'
$resultsDir = Join-Path $PSScriptRoot 'results'
New-Item -ItemType Directory -Path $resultsDir -Force | Out-Null

$resourceFile = Join-Path $resultsDir ("resources_{0}.csv" -f $stamp)
$summaryFile = Join-Path $resultsDir ("k6_summary_{0}.json" -f $stamp)
$reportFile = Join-Path $resultsDir ("run_summary_{0}.csv" -f $stamp)
$presetFile = Join-Path $resultsDir ("preset_delta_{0}.csv" -f $stamp)
$topFile = Join-Path $resultsDir ("postgres_top_{0}.csv" -f $stamp)
$deltaFile = Join-Path $resultsDir ("postgres_delta_{0}.csv" -f $stamp)
$countDeltaFile = Join-Path $resultsDir ("postgres_count_delta_{0}.csv" -f $stamp)
$scriptFile = Join-Path $PSScriptRoot 'prueba.js'
$includeTotal = if ($SkipExactTotal) { 'false' } else { 'true' }

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) { throw 'No se encontró docker en PATH.' }
if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) { throw 'No se encontró k6 en PATH.' }

function Invoke-DbQuery([string]$Sql) {
    $dockerArgs = @('exec', $PostgresContainer, 'psql', '-X', '-U', 'dating_app', '-d', 'dating_app', '-A', '-t', '-F', '|', '-c', $Sql)
    $output = & docker @dockerArgs
    if ($LASTEXITCODE -ne 0) { throw "Falló consulta de métricas PostgreSQL (exit $LASTEXITCODE)." }
    return $output
}

function Get-StatsResetAt {
    return (Invoke-DbQuery 'SELECT stats_reset::text FROM pg_stat_statements_info' | Select-Object -First 1).Trim()
}

function Get-PresetStats {
    $sql = @'
SELECT
    queryid::text,
    CASE WHEN position('LEFT JOIN profile_popularity pop' in query) > 0 THEN 'popular' ELSE 'recent' END,
    calls,
    rows,
    round(total_exec_time::numeric, 1),
    round(mean_exec_time::numeric, 1)
FROM pg_stat_statements
WHERE query LIKE '%p.id, p.display_name%'

  AND query LIKE '%ORDER BY%'
  AND query LIKE '%LIMIT $2%'
  AND query NOT LIKE '%p.id = ANY%'
ORDER BY 2, 1;
'@
    $lines = Invoke-DbQuery $sql
    $stats = @()
    foreach ($line in $lines) {
        if ([string]::IsNullOrWhiteSpace($line)) { continue }
        $fields = $line -split '\|'
        if ($fields.Count -lt 6) { continue }
        $stats += [pscustomobject]@{
            QueryId = $fields[0]
            Preset = $fields[1]
            Calls = [long]$fields[2]
            Rows = [long]$fields[3]
            TotalMs = [double]::Parse($fields[4], [System.Globalization.CultureInfo]::InvariantCulture)
            MeanMs = [double]::Parse($fields[5], [System.Globalization.CultureInfo]::InvariantCulture)
        }
    }
    return $stats
}

function Get-QueryStats {
    $sql = @'
SELECT
    queryid::text,
    calls,
    rows,
    round(total_exec_time::numeric, 1),
    replace(regexp_replace(query, E'[\n\r\t]+', ' ', 'g'), '|', '/')
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat%'
ORDER BY queryid;
'@
    $lines = Invoke-DbQuery $sql
    $stats = @()
    foreach ($line in $lines) {
        if ([string]::IsNullOrWhiteSpace($line)) { continue }
        $fields = $line -split '\|', 5
        if ($fields.Count -lt 5) { continue }
        $stats += [pscustomobject]@{
            QueryId = $fields[0]
            Calls = [long]$fields[1]
            Rows = [long]$fields[2]
            TotalMs = [double]::Parse($fields[3], [System.Globalization.CultureInfo]::InvariantCulture)
            Query = $fields[4]
        }
    }
    return $stats
}

if ($PrepareOnlineUsers) {
    Write-Host 'Marcando aleatoriamente el 30% de usuarios como activos durante los últimos 8 minutos.'
    $null = Invoke-DbQuery "UPDATE users SET last_active_at = now() - (random() * interval '8 minutes') WHERE random() < 0.30"
}

$statsResetBefore = Get-StatsResetAt
$presetBefore = @(Get-PresetStats)
$queryStatsBefore = @(Get-QueryStats)
$k6Version = (& k6 version | Select-Object -First 1).Trim()
if ($LASTEXITCODE -ne 0) { throw 'No se pudo consultar la versión de k6.' }

'Timestamp,Container,CPU_Percent,Memory_Usage' | Set-Content -LiteralPath $resourceFile -Encoding utf8
$job = Start-Job -ScriptBlock {
    param($file, $interval, $postgres, $backend)
    while ($true) {
        $timestamp = Get-Date -Format 'yyyy-MM-ddTHH:mm:ss'
        $lines = & docker stats --no-stream --format '{{.Name}},{{.CPUPerc}},{{.MemUsage}}' $postgres $backend
        foreach ($line in $lines) {
            Add-Content -LiteralPath $file -Value ("{0},{1}" -f $timestamp, $line) -Encoding utf8
        }
        Start-Sleep -Seconds $interval
    }
} -ArgumentList $resourceFile, $SampleIntervalSeconds, $PostgresContainer, $BackendContainer

$k6Args = @(
    'run', '--summary-export', $summaryFile,
    '-e', ("BASE_URL={0}" -f $BaseUrl),
    '-e', ("USERS={0}" -f $Users),
    '-e', ("MAX_PAGE={0}" -f $MaxPage),
    '-e', ("INCLUDE_TOTAL={0}" -f $includeTotal),
    '-e', ("DISCOVER_FILTER_MODE={0}" -f $DiscoverFilterMode),
    $scriptFile
)
$k6ExitCode = 1
try {
    Write-Host ("Ejecutando {0} con k6 {1}; resultados en {2}" -f $scriptFile, $k6Version, $resultsDir)
    & k6 @k6Args
    $k6ExitCode = $LASTEXITCODE
} finally {
    if ($job) {
        Stop-Job -Job $job -ErrorAction SilentlyContinue
        Receive-Job -Job $job -ErrorAction SilentlyContinue | Out-Null
        Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
    }
}

$statsResetAfter = Get-StatsResetAt
$presetAfter = @(Get-PresetStats)
$queryStatsAfter = @(Get-QueryStats)
if ($statsResetAfter -ne $statsResetBefore) {
    Write-Warning 'pg_stat_statements se reinició durante la prueba; los deltas de presets no son comparables.'
}

$beforeByKey = @{}
foreach ($row in $presetBefore) {
    $beforeByKey[('{0}:{1}' -f $row.QueryId, $row.Preset)] = $row
}
$deltas = foreach ($row in $presetAfter) {
    $key = '{0}:{1}' -f $row.QueryId, $row.Preset
    $old = $beforeByKey[$key]
    if (-not $old) { $old = [pscustomobject]@{ Calls = 0L; Rows = 0L; TotalMs = 0.0 } }
    $callsDelta = $row.Calls - $old.Calls
    $rowsDelta = $row.Rows - $old.Rows
    $totalDelta = [math]::Round($row.TotalMs - $old.TotalMs, 1)
    if ($callsDelta -gt 0 -or $rowsDelta -gt 0) {
        [pscustomobject]@{
            Preset = $row.Preset
            Calls = $callsDelta
            Rows = $rowsDelta
            TotalMs = $totalDelta
            MeanMs = if ($callsDelta -gt 0) { [math]::Round($totalDelta / $callsDelta, 1) } else { 0 }
            QueryId = $row.QueryId
        }
    }
}
if ($deltas) {
    $deltas | Export-Csv -LiteralPath $presetFile -NoTypeInformation -Encoding utf8
} else {
    'Preset,Calls,Rows,TotalMs,MeanMs,QueryId' | Set-Content -LiteralPath $presetFile -Encoding utf8
}

if (Test-Path -LiteralPath $summaryFile) {
    $summary = Get-Content -LiteralPath $summaryFile -Raw | ConvertFrom-Json
    $summary.PSObject.Properties.Remove('setup_data') | Out-Null
    $summary | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $summaryFile -Encoding utf8
    $rows = [System.Collections.Generic.List[object]]::new()
    $rows.Add([pscustomobject]@{ Metric = 'k6_version'; Value = $k6Version })
    $rows.Add([pscustomobject]@{ Metric = 'k6_exit_code'; Value = $k6ExitCode })
    $rows.Add([pscustomobject]@{ Metric = 'discover_filter_mode'; Value = $DiscoverFilterMode })
    $rows.Add([pscustomobject]@{ Metric = 'exact_total_mode'; Value = $includeTotal })
    $rows.Add([pscustomobject]@{ Metric = 'pg_stat_statements_reset_before'; Value = $statsResetBefore })
    $rows.Add([pscustomobject]@{ Metric = 'pg_stat_statements_reset_after'; Value = $statsResetAfter })
    foreach ($property in $summary.metrics.PSObject.Properties) {
        if ($property.Name -notlike 'http_req_duration*') { continue }
        $rows.Add([pscustomobject]@{ Metric = ($property.Name + '_p95_ms'); Value = $property.Value.'p(95)' })
        $rows.Add([pscustomobject]@{ Metric = ($property.Name + '_avg_ms'); Value = $property.Value.avg })
    }
    foreach ($metricName in @('http_reqs', 'vus_max')) {
        $metric = $summary.metrics.PSObject.Properties[$metricName]
        if ($metric) {
            foreach ($valueProperty in $metric.Value.PSObject.Properties) {
                $rows.Add([pscustomobject]@{ Metric = ($metricName + '_' + $valueProperty.Name); Value = $valueProperty.Value })
            }
        }
    }
    $failedRequests = $summary.metrics.PSObject.Properties['http_req_failed']
    if ($failedRequests) {
        $rows.Add([pscustomobject]@{ Metric = 'http_req_failed_rate'; Value = $failedRequests.Value.value })
        $rows.Add([pscustomobject]@{ Metric = 'http_req_failures'; Value = $failedRequests.Value.passes })
        $rows.Add([pscustomobject]@{ Metric = 'http_req_samples'; Value = ($failedRequests.Value.passes + $failedRequests.Value.fails) })
    }
    $rows | Export-Csv -LiteralPath $reportFile -NoTypeInformation -Encoding utf8
}

$topSql = "SELECT calls, round(mean_exec_time::numeric,1) AS mean_ms, round(total_exec_time::numeric,1) AS total_ms, rows, left(regexp_replace(query,'\s+',' ','g'),1500) AS query FROM pg_stat_statements WHERE query NOT LIKE '%pg_stat%' ORDER BY total_exec_time DESC LIMIT 20;"
$topOutput = & docker exec $PostgresContainer psql -X -U dating_app -d dating_app --csv -P pager=off -c $topSql
if ($LASTEXITCODE -ne 0) { throw 'Falló la extracción del top de PostgreSQL.' }
$topOutput | Set-Content -LiteralPath $topFile -Encoding utf8

$queryBeforeById = @{}
foreach ($row in $queryStatsBefore) { $queryBeforeById[$row.QueryId] = $row }
$queryDeltas = @()
if ($statsResetAfter -eq $statsResetBefore) {
    foreach ($row in $queryStatsAfter) {
        $old = $queryBeforeById[$row.QueryId]
        if (-not $old) { $old = [pscustomobject]@{ Calls = 0L; Rows = 0L; TotalMs = 0.0 } }
        $callsDelta = $row.Calls - $old.Calls
        $rowsDelta = $row.Rows - $old.Rows
        $totalDelta = [math]::Round($row.TotalMs - $old.TotalMs, 1)
        if ($callsDelta -gt 0 -and $rowsDelta -ge 0 -and $totalDelta -ge 0) {
            $queryDeltas += [pscustomobject]@{
                Calls = $callsDelta
                MeanMs = [math]::Round($totalDelta / $callsDelta, 1)
                TotalMs = $totalDelta
                Rows = $rowsDelta
                QueryId = $row.QueryId
                Query = $row.Query
            }
        }
    }
    if ($queryDeltas.Count -gt 0) {
        $queryDeltas | Sort-Object TotalMs -Descending | Select-Object -First 20 |
            Export-Csv -LiteralPath $deltaFile -NoTypeInformation -Encoding utf8
    } else {
        'Calls,MeanMs,TotalMs,Rows,QueryId,Query' | Set-Content -LiteralPath $deltaFile -Encoding utf8
    }
    $countDeltas = @($queryDeltas | Where-Object { $_.Query -match '^SELECT COUNT\(\*\)' })
    if ($countDeltas.Count -gt 0) {
        $countDeltas | Sort-Object TotalMs -Descending |
            Export-Csv -LiteralPath $countDeltaFile -NoTypeInformation -Encoding utf8
    } else {
        'Calls,MeanMs,TotalMs,Rows,QueryId,Query' | Set-Content -LiteralPath $countDeltaFile -Encoding utf8
    }
} else {
    Write-Warning 'pg_stat_statements se reinició durante la prueba; no se puede calcular el delta SQL.'
    'Calls,MeanMs,TotalMs,Rows,QueryId,Query' | Set-Content -LiteralPath $deltaFile -Encoding utf8
    'Calls,MeanMs,TotalMs,Rows,QueryId,Query' | Set-Content -LiteralPath $countDeltaFile -Encoding utf8
}

Write-Host ("Prueba terminada. Código k6: {0}" -f $k6ExitCode)
Write-Host ("Informe: {0}" -f $reportFile)
Write-Host ("Deltas de presets: {0}" -f $presetFile)
Write-Host ("CPU/RAM: {0}" -f $resourceFile)
Write-Host ("Top PostgreSQL acumulado: {0}" -f $topFile)
Write-Host ("Deltas SQL PostgreSQL: {0}" -f $deltaFile)
Write-Host ("Deltas de COUNT(*) PostgreSQL: {0}" -f $countDeltaFile)
Write-Host ("Resumen k6 original: {0}" -f $summaryFile)
if ($k6ExitCode -ne 0) { exit $k6ExitCode }
