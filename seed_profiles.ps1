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

$interesesList = @(
    @("viajes", "fotografía", "café"),
    @("música", "conciertos", "cine"),
    @("senderismo", "perros", "naturaleza"),
    @("cocina", "gastronomía", "vinos"),
    @("tecnología", "videojuegos", "lectura"),
    @("arte", "museos", "diseño"),
    @("fitness", "yoga", "nutrición"),
    @("playa", "surf", "camping")
)

$goals = @("long_term", "short_term", "casual", "marriage")

Write-Host "=== Iniciando la creación de 50 perfiles reales ===" -ForegroundColor Green

for ($i = 0; $i -lt 50; $i++) {
    $num = $i + 1
    $email = "user$num@datingdemo.com"
    $pass  = "DemoPass123!"
    $nombre = $nombres[$i]
    $genero = $generos[$i]
    $bio    = $bios[$i % $bios.Count]
    $intereses = $interesesList[$i % $interesesList.Count]
    $goal   = $goals[$i % $goals.Count]
    
    # Edad entre 20 y 38 años
    $year = 2004 - ($i % 18)
    $month = (($i % 12) + 1).ToString("00")
    $day = (($i % 27) + 1).ToString("00")
    $birthDate = "$year-$month-$day"

    Write-Host "[$num/50] Creando a $nombre ($email)..." -NoNewline

    # 1. Registro
    $bodyReg = @{
        email = $email
        password = $pass
        accepted_terms = $true
    } | ConvertTo-Json

    try {
        $sess = New-Object Microsoft.PowerShell.Commands.WebRequestSession
        $regRes = Invoke-RestMethod -Uri "http://localhost/api/v1/auth/register" -Method Post -Body $bodyReg -ContentType "application/json" -SessionVariable sess
    } catch {
        # Si ya existe, hacemos login
        $bodyLogin = @{ email = $email; password = $pass } | ConvertTo-Json
        $loginRes = Invoke-RestMethod -Uri "http://localhost/api/v1/auth/login" -Method Post -Body $bodyLogin -ContentType "application/json" -SessionVariable sess
    }

    # 2. Crear Perfil
    $bodyProfile = @{
        display_name = $nombre
        birth_date = $birthDate
        gender = $genero
        country_code = "ES"
        languages = @("es", "en")
        relationship_goal = $goal
        bio = $bio
        interests = $intereses
    } | ConvertTo-Json

    try {
        $profileRes = Invoke-RestMethod -Uri "http://localhost/api/v1/profiles/me" -Method Post -Body $bodyProfile -ContentType "application/json" -WebSession $sess
    } catch {
        # Si ya tenía perfil creado, seguimos
    }

    # 3. Subir Foto (rotando las 5 fotos reales)
    $fotoSeleccionada = $fotos[$i % $fotos.Count]
    $cookieVal = $sess.Cookies.GetCookies("http://localhost")["session_id"].Value

    $subida = curl.exe -s -b "session_id=$cookieVal" -X POST "http://localhost/api/v1/profiles/me/photos" -F "photo=@$fotoSeleccionada;type=image/jpeg"

    Write-Host " [OK]" -ForegroundColor Green
}

Write-Host "=== ¡Listo! 50 perfiles creados y poblados con fotos reales ===" -ForegroundColor Cyan
