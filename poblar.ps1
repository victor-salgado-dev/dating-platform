Write-Host "=== Limpiando contador de Rate Limit en Redis ===" -ForegroundColor Cyan
docker compose exec redis redis-cli flushall | Out-Null

$fotosDir = "C:\Users\Victor\Downloads\fotos\Nueva carpeta (4)"
$fotos = Get-ChildItem -Path $fotosDir -Filter "*.jpg" | Select-Object -ExpandProperty FullName

if ($fotos.Count -eq 0) {
    Write-Host "Error: No se encontraron fotos en $fotosDir" -ForegroundColor Red
    exit 1
}

$nombres = @("Lucía","Sofía","Martina","María","Paula","Julia","Emma","Valeria","Daniela","Alba",
             "Mateo","Hugo","Martín","Lucas","Leo","Daniel","Alejandro","Manuel","Pablo","Álvaro",
             "Elena","Carmen","Sara","Victoria","Clara","Mario","Adrián","Diego","David","Javier",
             "Laura","Carla","Irene","Marina","Rocío","Carlos","Gonzalo","Marcos","Sergio","Jorge",
             "Natalia","Celia","Nuria","Ainhoa","Blanca","Iván","Rubén","Raúl","Víctor","Óscar")

$generos = @("female","female","female","female","female","female","female","female","female","female",
             "male","male","male","male","male","male","male","male","male","male",
             "female","female","female","female","female","male","male","male","male","male",
             "female","female","female","female","female","male","male","male","male","male",
             "female","female","female","female","female","male","male","male","male","male")

$regiones = @("Madrid", "Barcelona", "Valencia", "Sevilla", "Zaragoza", "Málaga", "Bilbao", "Alicante", "Granada", "Vigo")

$bios = @(
    "Apasionada del café, los viajes espontáneos y los libros que no puedes soltar.",
    "Buscando a alguien con quien compartir risas, conciertos y buenas conversaciones.",
    "Amante de la naturaleza, el senderismo de fin de semana y las noches de cine.",
    "Ingeniero de día, cocinero aficionado de noche. Siempre listo para probar cosas nuevas.",
    "Me encanta el arte, los museos tranquilos y perderse por ciudades desconocidas.",
    "Adicto al deporte, los perros y una buena charla con cerveza artesanal.",
    "Diseñadora gráfica viviendo entre colores, música indie y buena vibra.",
    "Buscando conexión genuina. Menos superficialidad y más historias reales.",
    "Si te gusta la pizza con piña, tendremos que debatirlo seriamente.",
    "Curioso por naturaleza, siempre aprendiendo algo nuevo o planeando una escapada."
)

$interesesCatalog = @(
    @("viajes", "fotografia_urbana", "cafes"),
    @("musica", "conciertos_festivales", "cine_series"),
    @("senderismo", "perros", "naturaleza"),
    @("cocina", "salir_a_comer", "cocina_italiana"),
    @("tecnologia", "videojuegos", "lectura"),
    @("arte_y_cultura", "dibujo", "lectura"),
    @("fitness_gimnasio", "yoga_pilates", "comida_saludable"),
    @("viajes_de_playa", "surf", "camping")
)

$goals = @("long_term", "short_term", "casual", "marriage")

Write-Host "=== Creando 50 perfiles con pausas de seguridad ===" -ForegroundColor Green

$createdProfiles = @()

for ($i = 0; $i -lt 50; $i++) {
    $num = $i + 1
    $email = "user$num@datingdemo.com"
    $pass  = "DemoPass123!"
    $nombre = $nombres[$i]
    $genero = $generos[$i]
    $bio    = $bios[$i % $bios.Count]
    $intereses = $interesesCatalog[$i % $interesesCatalog.Count]
    $goal   = $goals[$i % $goals.Count]
    $region = $regiones[$i % $regiones.Count]

    $year = 2004 - ($i % 18)
    $month = (($i % 12) + 1).ToString("00")
    $day = (($i % 27) + 1).ToString("00")
    $birthDate = "$year-$month-$day"

    Write-Host "[$num/50] $nombre ($email)... " -NoNewline

    $bodyReg = @{
        email = $email
        password = $pass
        accepted_terms = $true
    } | ConvertTo-Json

    $sess = New-Object Microsoft.PowerShell.Commands.WebRequestSession
    $authed = $false

    while (-not $authed) {
        try {
            $null = Invoke-RestMethod -Uri "http://localhost/api/v1/auth/register" -Method Post -Body $bodyReg -ContentType "application/json" -SessionVariable sess
            $authed = $true
        } catch {
            if ($_.Exception.Response.StatusCode.value__ -eq 429) {
                Write-Host "(rate limit, esperando)... " -NoNewline -ForegroundColor Yellow
                Start-Sleep -Seconds 2
            } else {
                try {
                    $bodyLogin = @{ email = $email; password = $pass } | ConvertTo-Json
                    $null = Invoke-RestMethod -Uri "http://localhost/api/v1/auth/login" -Method Post -Body $bodyLogin -ContentType "application/json" -SessionVariable sess
                    $authed = $true
                } catch {
                    if ($_.Exception.Response.StatusCode.value__ -eq 429) {
                        Start-Sleep -Seconds 2
                    } else {
                        $authed = $true
                    }
                }
            }
        }
    }

    Start-Sleep -Milliseconds 300

    $bodyProfile = @{
        display_name = $nombre
        birth_date = $birthDate
        gender = $genero
        country_code = "ES"
        region = $region
        relationship_goals = @($goal)
        bio = $bio
    } | ConvertTo-Json

    $profileId = $null
    try {
        $profileRes = Invoke-RestMethod -Uri "http://localhost/api/v1/profiles/me" -Method Post -Body $bodyProfile -ContentType "application/json" -WebSession $sess
        $profileId = $profileRes.id
    } catch {
        try {
            $myProfile = Invoke-RestMethod -Uri "http://localhost/api/v1/profiles/me" -Method Get -WebSession $sess
            $profileId = $myProfile.id
        } catch {}
    }

    if ($profileId) {
        $createdProfiles += @{ id = $profileId; sess = $sess; name = $nombre }
    }

    foreach ($intKey in $intereses) {
        try {
            $bodyInt = @{ level = 3 } | ConvertTo-Json
            Invoke-RestMethod -Uri "http://localhost/api/v1/profiles/me/interests/$intKey" -Method Put -Body $bodyInt -ContentType "application/json" -WebSession $sess | Out-Null
        } catch {}
    }

    $cookieObj = $sess.Cookies.GetCookies("http://localhost")["session_id"]
    if ($cookieObj) {
        $cookieVal = $cookieObj.Value
        for ($f = 0; $f -lt 3; $f++) {
            $fotoIndex = ($i * 3 + $f) % $fotos.Count
            $fotoSeleccionada = $fotos[$fotoIndex]
            curl.exe -s -b "session_id=$cookieVal" -X POST "http://localhost/api/v1/profiles/me/photos" -F "photo=@$fotoSeleccionada;type=image/jpeg" | Out-Null
        }
    }

    Write-Host "[OK - 3 fotos]" -ForegroundColor Green
    Start-Sleep -Milliseconds 200
}

Write-Host "=== Generando visitas entre perfiles ===" -ForegroundColor Yellow
for ($i = 0; $i -lt [Math]::Min(25, $createdProfiles.Count); $i++) {
    $visitor = $createdProfiles[$i]
    for ($v = 1; $v -le 4; $v++) {
        $targetIndex = ($i + $v * 2) % $createdProfiles.Count
        $target = $createdProfiles[$targetIndex]
        if ($target.id -ne $visitor.id) {
            try {
                Invoke-RestMethod -Uri "http://localhost/api/v1/visits/$($target.id)" -Method Post -WebSession $visitor.sess | Out-Null
            } catch {}
        }
    }
    Start-Sleep -Milliseconds 100
}

Write-Host "=== ¡Listo! Base de datos poblada sin bloqueos ===" -ForegroundColor Cyan
