$csvFile = ".\metricas.csv"
$jsonFile = ".\k6_resumen.json"

# 1. Resetear estadísticas de Postgres
Write-Host "Reseteando estadísticas de PostgreSQL..." -ForegroundColor Cyan
docker exec dating-platform-postgres-1 psql -U dating_app -d dating_app -c "SELECT pg_stat_statements_reset();" | Out-Null

# 2. Despertar a ~1.500 usuarios para que 'online-now' SIEMPRE tenga perfiles activos
Write-Host "Simulando actividad reciente de usuarios (Online Now)..." -ForegroundColor Cyan
docker exec dating-platform-postgres-1 psql -U dating_app -d dating_app -c "UPDATE users SET last_active_at = now() - (random() * interval '8 minutes') WHERE random() < 0.30;" | Out-Null

# 3. Iniciar CSV con cabecera de Docker
"Fecha,Hora,Contenedor,CPU_Porcentaje,Memoria_Uso" | Out-File -FilePath $csvFile -Encoding utf8
Write-Host "Iniciando monitor de CPU/RAM..." -ForegroundColor Cyan

# 4. Monitor de Docker en segundo plano
$job = Start-Job -ScriptBlock {
    param($file)
    while ($true) {
        $timestamp = Get-Date -Format "yyyy-MM-dd,HH:mm:ss"
        $stats = docker stats --no-stream --format "{{.Name}},{{.CPUPerc}},{{.MemUsage}}" | Select-String "backend|postgres"
        foreach ($line in $stats) {
            "$timestamp,$line" | Out-File -FilePath $file -Append -Encoding utf8
        }
        Start-Sleep -Seconds 3
    }
} -ArgumentList (Resolve-Path $csvFile)

# 5. Lanzar test de k6
Write-Host "Lanzando test de k6..." -ForegroundColor Green
k6 run --summary-export $jsonFile prueba.js

# 6. Parar monitor
Write-Host "Test finalizado. Deteniendo monitor..." -ForegroundColor Cyan
Stop-Job $job
Remove-Job $job

# 7. Extraer Métricas de k6 (Tiempos y Latencias)
if (Test-Path $jsonFile) {
    try {
        $raw = Get-Content $jsonFile -Raw | ConvertFrom-Json
        $m = $raw.metrics
        
        $dur = $m.http_req_duration
        $req = $m.http_reqs
        $vus = $m.vus_max.value
        if (-not $vus) { $vus = $m.vus.value }
        $fail = $m.http_req_failed.value

        "" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "# ========================================================" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "# RESUMEN K6: TIEMPOS DE RESPUESTA Y LATENCIAS" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "# ========================================================" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Metrica,Valor" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Usuarios_Maximos_VUs,$vus" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Peticiones_Totales,$($req.count)" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Peticiones_Por_Segundo_RPS,$([math]::Round($req.rate, 2))" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Latencia_Media,$([math]::Round($dur.avg, 2)) ms" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Latencia_Mediana,$([math]::Round($dur.med, 2)) ms" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Latencia_P90,$([math]::Round($dur.'p(90)', 2)) ms" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Latencia_P95,$([math]::Round($dur.'p(95)', 2)) ms" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Latencia_Maxima,$([math]::Round($dur.max, 2)) ms" | Out-File -FilePath $csvFile -Append -Encoding utf8
        "Porcentaje_Fallos,$([math]::Round($fail * 100, 2))%" | Out-File -FilePath $csvFile -Append -Encoding utf8

        Remove-Item $jsonFile -ErrorAction SilentlyContinue
    } catch {}
}

# 8. Top 5 de PostgreSQL exacto
"" | Out-File -FilePath $csvFile -Append -Encoding utf8
"# ========================================================" | Out-File -FilePath $csvFile -Append -Encoding utf8
"# TOP 5 CONSULTAS POSTGRESQL (TIEMPO REAL DE BASE DE DATOS)" | Out-File -FilePath $csvFile -Append -Encoding utf8
"# ========================================================" | Out-File -FilePath $csvFile -Append -Encoding utf8

$sqlQuery = "SELECT calls, round(mean_exec_time::numeric,1) AS mean_ms, round(total_exec_time::numeric) AS total_ms, left(regexp_replace(query,'\s+',' ','g'),1500) AS query FROM pg_stat_statements WHERE query NOT LIKE '%pg_stat%' ORDER BY total_exec_time DESC LIMIT 5;"
$pgResults = docker exec dating-platform-postgres-1 psql -U dating_app -d dating_app --csv -P pager=off -c $sqlQuery
foreach ($line in $pgResults) {
    if ($line.Trim() -ne "") {
        $line | Out-File -FilePath $csvFile -Append -Encoding utf8
    }
}

Write-Host "`n¡Informe completo generado en: metricas.csv!" -ForegroundColor Yellow
