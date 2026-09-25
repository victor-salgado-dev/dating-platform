# seed_profiles.ps1 - Versión con Bypass de Rate Limit + Auto-Wait + Campos Completos
$baseUrl    = "http://localhost"
$totalUsers = 500
$delayMs    = 0
$cookieName = "session_id"
$fotosDir   = "C:\Users\Victor\Downloads\fotos\Nueva carpeta (4)"

# --- Funciones de Gestión de Rate Limit (Redis) -----------------------------
function Reset-RedisBucket {
    try {
        & docker compose exec -T redis redis-cli FLUSHALL 2>$null | Out-Null
    } catch {}
}

function Clear-RateLimits {
    try {
        & docker compose exec -T redis redis-cli EVAL "local k = redis.call('keys', '*rate*'); for i=1,#k do redis.call('del', k[i]) end; local k2 = redis.call('keys', '*limit*'); for i=1,#k2 do redis.call('del', k2[i]) end;" 0 2>$null | Out-Null
    } catch {}
}

# --- 0. Esperar a que el backend esté listo ----------------------------------
Write-Host "=== Comprobando disponibilidad del backend ===" -ForegroundColor Cyan
$backendReady = $false
for ($w = 0; $w -lt 60; $w++) {
    try {
        $res = Invoke-WebRequest -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body "{}" -ContentType "application/json" -ErrorAction Stop
        $backendReady = $true
        break
    } catch {
        if ($_.Exception.Response.StatusCode -ne 502 -and $_.Exception.Response.StatusCode -ne $null) {
            $backendReady = $true
            break
        }
    }
    Write-Host "Esperando al backend (~30s si acaba de arrancar Docker)... ($w)" -ForegroundColor Yellow
    Start-Sleep -Seconds 2
}

if (-not $backendReady) {
    Write-Host "El backend no responde tras 60 segundos. Aborta." -ForegroundColor Red
    exit
}
Reset-RedisBucket
Write-Host "¡Backend Listo! Comenzando generación de 500 usuarios..." -ForegroundColor Green

# --- Fotos disponibles ------------------------------------------------------
$fotos = @()
if (Test-Path $fotosDir) {
    $fotos = @(Get-ChildItem -Path $fotosDir -Filter "*.jpg" | Select-Object -ExpandProperty FullName)
}
if ($fotos.Count -eq 0) {
    Write-Host "Aviso: no se encontraron .jpg en $fotosDir - se crearán perfiles sin foto." -ForegroundColor Yellow
}

# --- Catálogos y datos base -------------------------------------------------
$nombres = @("Lucia","Sofia","Martina","Maria","Paula","Julia","Emma","Valeria","Daniela","Alba",
             "Mateo","Hugo","Martin","Lucas","Leo","Daniel","Alejandro","Manuel","Pablo","Alvaro",
             "Elena","Carmen","Sara","Victoria","Clara","Mario","Adrian","Diego","David","Javier",
             "Laura","Carla","Irene","Marina","Rocio","Carlos","Gonzalo","Marcos","Sergio","Jorge")

$generos = @("female","female","female","female","female","male","male","male","male","male","non_binary","other")

$bios = @(
    "Apasionada del café, los viajes espontáneos y los libros que no puedes soltar.",
    "Buscando a alguien con quien compartir risas, conciertos y buenas conversaciones.",
    "Amante de la naturaleza, el senderismo de fin de semana y las noches de cine.",
    "Ingeniero de día, cocinero aficionado de noche. Siempre listo para probar cosas nuevas.",
    "Me encanta el arte, los museos tranquilos y perderse por ciudades desconocidas.",
    "Adicto al deporte, los perros y una buena charla con cerveza artesanal.",
    "Diseñadora gráfica viviendo entre colores, música indie y buena vibra.",
    "Buscando conexión genuina. Menos superficialidad y más historias reales."
)

$regiones = @("Madrid","Barcelona","Valencia","Sevilla","Bilbao","Málaga","Zaragoza","Murcia","Alicante","Granada")

# Opciones de los campos multi-select y simples de Perfil (Migraciones 12 a 16)
$hasChildrenOptions     = @("yes","no","prefer_not_to_say")
$wantsChildrenOptions   = @("yes","no","not_sure")
$relationshipGoalOptions= @("casual","long_term","friendship","marriage","not_sure")
$bodyTypes              = @("petite","slim","athletic","average","few_extra_pounds","full_figured","large_and_lovely")
$ethnicities            = @("arab","asian","black","caucasian","hispanic","indian","mixed","pacific_islander","other")
$appearanceRatings      = @("below_average","average","attractive","very_attractive")
$hairColors             = @("bald","black","blonde","brown","grey","light_brown","red","changes_frequently","other")
$eyeColors              = @("black","blue","brown","green","grey","hazel","other")
$bodyArtOptions         = @("branding","earrings","piercing","tattoo","other")
$smokeOptions           = @("yes","no","occasionally")
$drinkOptions           = @("yes","no","occasionally")
$relocOptions           = @("within_country","another_country","not_willing","not_sure")
$maritalStatuses        = @("single","separated","widowed","divorced","other")
$employmentStatuses     = @("student","part_time","full_time","homemaker","retired","not_employed","other")
$incomeLevels           = @("low","medium","high","very_high","prefer_not_to_say")
$livingSituations       = @("live_alone","live_with_friends","live_with_family","live_with_kids","live_with_spouse","other")
$educationLevels        = @("high_school","associates","bachelors","masters","phd","other")
$englishAbilities       = @("none","basic","intermediate","fluent","native")
$religions              = @("bahai","buddhist","catholic","christian_other","protestant","hindu","islam","jainism","jewish","parsi","shintoism","sikhism","taoism","other","none")
$religiousValues        = @("not_religious","religious","very_religious")
$starSigns              = @("aquarius","aries","cancer","capricorn","gemini","leo","libra","pisces","sagittarius","scorpio","taurus","virgo")
$futureVisionOptions    = @("balance_family_career","focus_family_household","beauty_and_partner_time","part_time_work","support_partner_career","new_education")
$sportsOptions          = @("fitness","motorsport","strength_training","ball_sports","water_sports","jogging","winter_sports","cycling","athletics","climbing","horse_riding","hiking","other")
$likesPetsOptions       = @("yes","neutral","no")
$petsOwnedOptions       = @("cat","dog","horse","other")
$seasons                = @("spring","summer","autumn","winter")
$vacationStyles         = @("small_charming_hotel","luxury_hotel","cruise_ship","club_hotel","rental_apartment","countryside_house","camping_rv","staying_home","at_friends")
$vacationActivitiesOpts = @("cafes_shopping_nightlife","mix_relaxation_activities","lazing_and_relaxing","beach_holiday","sightseeing_cities","lots_of_sports")
$occupations = @("administrative","advertising","artistic","construction","education","entertainment","finance","medical","military","retail","sales","self_employed","student","technical","travel","other")
$quotes = @("La vida es eso que pasa mientras haces otros planes.", "Colecciona momentos, no cosas.", "Ríe fuerte, ama sin miedo.")
$dreams = @("Recorrer Japón en tren.", "Aprender a tocar el piano.", "Vivir un año en otra ciudad.", "Abrir un pequeño café-librería.")

$languageCodes = @("es","en","fr","de","it","pt","ja","zh_cmn","ar","ru")
$personalityKeys = @(
    "extra_reserved_calm","extra_funny_laughs","extra_center_of_party","extra_enjoys_alone_time",
    "emo_sensitive_vulnerable","emo_moody","emo_self_confident","emo_hard_to_rattle",
    "cons_chaotic","cons_no_planning","cons_goal_oriented","cons_very_tidy",
    "agree_distrustful_at_first","agree_helpful_caring","agree_hard_to_get_along","agree_believes_in_good",
    "open_original_new_ideas","open_cautious_with_new","open_interested_arts","open_traditions_matter"
)

$interestsWithLevel = @("deporte_actividad_fisica", "fitness_gimnasio", "running", "ciclismo", "bricolaje_diy", "fotografia", "cocina", "lectura", "tecnologia_informatica", "musica", "cine_series", "videojuegos", "viajar", "naturaleza", "socializar")
$interestsNoLevel   = @("dibujo", "poesia", "bricolaje", "cantar", "guitarra", "pop", "rock", "indie", "pc_gaming", "anime", "ciencia_ficcion", "futbol", "tenis", "coches", "perros", "gatos", "viajes", "cocina_italiana", "sushi", "accion", "comedia", "novelas")

$desiredTraitsOptions = @("humorous","self_confident","loving","kind_hearted","intelligent","faithful","honest","ambitious","family_oriented","adventurous","romantic")
$partnerMayHaveChildrenOptions = @("yes","no","doesnt_matter")
$partnerReligionOptions        = $religions + @("doesnt_matter")
$firstMeetingOptions           = @("doesnt_matter","public_place","my_city","their_city","video_call_first")
$desiredLivingPlaceOptions     = @("big_city","medium_city","small_town","countryside","abroad")

$profileIdMap = @{}

# Función auxiliar para login/registro que maneja rate limit automáticamente
function Get-AuthSession($num, $email, $pass) {
    $ip = "192.168.1.$($num % 250 + 1)"
    $headers = @{ "X-Forwarded-For" = $ip; "X-Real-IP" = $ip }

    for ($att = 1; $att -le 5; $att++) {
        $sess = New-Object Microsoft.PowerShell.Commands.WebRequestSession
        $bodyReg = @{ email = $email; password = $pass; accepted_terms = $true } | ConvertTo-Json
        try {
            $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" -Method Post `
                -Body $bodyReg -ContentType "application/json" -Headers $headers -WebSession $sess
            return $sess
        } catch {
            $err = $_.Exception.Message
            if ($err -match "429" -or $_.ErrorDetails.Message -match "rate_limited") {
                Reset-RedisBucket
                Start-Sleep -Milliseconds 200
                continue
            }
            # Si el usuario ya existía, iniciamos sesión
            $bodyLogin = @{ email = $email; password = $pass } | ConvertTo-Json
            try {
                $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post `
                    -Body $bodyLogin -ContentType "application/json" -Headers $headers -WebSession $sess
                return $sess
            } catch {
                if ($_.Exception.Message -match "429" -or $_.ErrorDetails.Message -match "rate_limited") {
                    Reset-RedisBucket
                    Start-Sleep -Milliseconds 200
                    continue
                }
            }
        }
    }
    return $null
}

Write-Host "=== 1/2 Creando $totalUsers perfiles demo ===" -ForegroundColor Green

for ($i = 0; $i -lt $totalUsers; $i++) {
    $num       = $i + 1
    $email     = "user$num@datingdemo.com"
    $pass      = "DemoPass123!"
    $nombre    = $nombres | Get-Random
    $genero    = $generos | Get-Random
    $bio       = $bios | Get-Random

    $year      = Get-Random -Minimum 1975 -Maximum 2005
    $month     = (Get-Random -Minimum 1 -Maximum 13).ToString("00")
    $day       = (Get-Random -Minimum 1 -Maximum 28).ToString("00")
    $birthDate = "$year-$month-$day"

    # Prevenir rate limiting vaciando contadores de Redis periódicamente
    if ($i % 8 -eq 0) {
        Reset-RedisBucket
    }

    Write-Host ("[{0}/{1}] {2} ({3})..." -f $num, $totalUsers, $nombre, $email) -NoNewline

    # 1) Auth con bypass de Rate Limiting
    $sess = Get-AuthSession $num $email $pass
    if (-not $sess) {
        Write-Host " [ERROR AUTH]" -ForegroundColor Red
        continue
    }

    # 2) Completar Perfil Extendido (Migración 12, 13, 15)
    $hasChildren = $hasChildrenOptions | Get-Random
    
    $profileBody = [ordered]@{
        display_name           = $nombre
        birth_date             = $birthDate
        gender                 = $genero
        country_code           = "ES"
        region                 = $regiones | Get-Random
        relationship_goals     = @($relationshipGoalOptions | Get-Random -Count (Get-Random -Minimum 1 -Maximum 3))
        has_children           = $hasChildren
        wants_children         = $wantsChildrenOptions | Get-Random
        bio                    = $bio
        height                 = Get-Random -Minimum 150 -Maximum 205
        weight                 = Get-Random -Minimum 50 -Maximum 110
        body_type              = $bodyTypes | Get-Random
        ethnicity              = $ethnicities | Get-Random
        appearance_rating      = $appearanceRatings | Get-Random
        hair_color             = $hairColors | Get-Random
        eye_color              = $eyeColors | Get-Random
        body_art               = @($bodyArtOptions | Get-Random -Count (Get-Random -Minimum 1 -Maximum 3))
        smoking_habit          = $smokeOptions | Get-Random
        drinking_habit         = $drinkOptions | Get-Random
        relocation_willingness = @($relocOptions | Get-Random -Count (Get-Random -Minimum 1 -Maximum 3))
        marital_status         = $maritalStatuses | Get-Random
        occupation             = $occupations | Get-Random
        employment_status      = $employmentStatuses | Get-Random
        income_level           = $incomeLevels | Get-Random
        living_situation       = $livingSituations | Get-Random
        nationality            = "ES"
        education_level        = $educationLevels | Get-Random
        english_ability        = $englishAbilities | Get-Random
        religion               = $religions | Get-Random
        religious_values       = $religiousValues | Get-Random
        star_sign              = $starSigns | Get-Random
        future_vision          = @($futureVisionOptions | Get-Random -Count (Get-Random -Minimum 1 -Maximum 3))
        sports                 = @($sportsOptions | Get-Random -Count (Get-Random -Minimum 1 -Maximum 4))
        likes_pets             = $likesPetsOptions | Get-Random
        pets_owned             = @($petsOwnedOptions | Get-Random -Count (Get-Random -Minimum 1 -Maximum 3))
        favorite_season        = $seasons | Get-Random
        ideal_vacation_style   = @($vacationStyles | Get-Random -Count (Get-Random -Minimum 1 -Maximum 3))
        vacation_activities    = @($vacationActivitiesOpts | Get-Random -Count (Get-Random -Minimum 1 -Maximum 4))
        profile_quote          = $quotes | Get-Random
        dream_wish             = $dreams | Get-Random
    }

    if ($hasChildren -eq 'yes') {
        $profileBody.children_count     = Get-Random -Minimum 1 -Maximum 4
        $profileBody.youngest_child_age = Get-Random -Minimum 0 -Maximum 10
        $profileBody.oldest_child_age   = Get-Random -Minimum 11 -Maximum 18
    }

    $profileJson = $profileBody | ConvertTo-Json -Depth 10
    $createdProfile = $null
    try {
        $createdProfile = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me" -Method Post `
            -Body $profileJson -ContentType "application/json" -WebSession $sess
    } catch {
        try {
            $createdProfile = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me" -Method Get -WebSession $sess
        } catch {}
    }

    $currentProfileId = if ($createdProfile.data.id) { $createdProfile.data.id } else { $createdProfile.id }
    if ($currentProfileId) { $profileIdMap[$num] = $currentProfileId }

    # 3) Idiomas (Migración 14)
    $langsToSet = @($languageCodes | Get-Random -Count (Get-Random -Minimum 1 -Maximum 4))
    foreach ($code in $langsToSet) {
        $langJson = @{ level = Get-Random -Minimum 1 -Maximum 6 } | ConvertTo-Json
        try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/languages/$code" -Method Put -Body $langJson -ContentType "application/json" -WebSession $sess } catch {}
    }

    # 4) Personalidad (Migración 13)
    foreach ($pk in $personalityKeys) {
        $persJson = @{ score = Get-Random -Minimum 1 -Maximum 6 } | ConvertTo-Json
        try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/personality/$pk" -Method Put -Body $persJson -ContentType "application/json" -WebSession $sess } catch {}
    }

    # 5) Preferencias de Pareja (Migración 13)
    $aMin = Get-Random -Minimum 18 -Maximum 35
    $hMin = Get-Random -Minimum 150 -Maximum 170
    $partnerBody = [ordered]@{
        age_min    = $aMin
        age_max    = $aMin + (Get-Random -Minimum 5 -Maximum 20)
        height_min = $hMin
        height_max = $hMin + (Get-Random -Minimum 10 -Maximum 30)
        desired_traits = @($desiredTraitsOptions | Get-Random -Count (Get-Random -Minimum 2 -Maximum 5))
        partner_may_have_children = $partnerMayHaveChildrenOptions | Get-Random
        partner_religion_preference = $partnerReligionOptions | Get-Random
        about_partner_text = "Busco a alguien con quien compartir momentos geniales y viajar."
        first_meeting_preference = $firstMeetingOptions | Get-Random
        desired_living_place = @($desiredLivingPlaceOptions | Get-Random -Count (Get-Random -Minimum 1 -Maximum 3))
        
        importance_shared_thoughts     = Get-Random -Minimum 1 -Maximum 6
        importance_shared_hobbies      = Get-Random -Minimum 1 -Maximum 6
        importance_intimacy            = Get-Random -Minimum 1 -Maximum 6
        importance_romantic_love       = Get-Random -Minimum 1 -Maximum 6
        importance_financial_security  = Get-Random -Minimum 1 -Maximum 6
        importance_fun                 = Get-Random -Minimum 1 -Maximum 6
        importance_shared_friends      = Get-Random -Minimum 1 -Maximum 6
        importance_shared_humor        = Get-Random -Minimum 1 -Maximum 6
        importance_personal_space      = Get-Random -Minimum 1 -Maximum 6
        importance_independence        = Get-Random -Minimum 1 -Maximum 6
    } | ConvertTo-Json -Depth 10

    try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/partner-preferences" -Method Patch -Body $partnerBody -ContentType "application/json" -WebSession $sess } catch {}

    # 6) Intereses / Hobbies (Migración 16)
    $selIntW = @($interestsWithLevel | Get-Random -Count (Get-Random -Minimum 2 -Maximum 5))
    foreach ($intW in $selIntW) {
        $intJson = @{ level = (Get-Random -Minimum 1 -Maximum 6) } | ConvertTo-Json
        try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/interests/$intW" -Method Put -Body $intJson -ContentType "application/json" -WebSession $sess } catch {}
    }

    $selIntN = @($interestsNoLevel | Get-Random -Count (Get-Random -Minimum 2 -Maximum 6))
    foreach ($intN in $selIntN) {
        $intJson2 = "{}"
        try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/interests/$intN" -Method Put -Body $intJson2 -ContentType "application/json" -WebSession $sess } catch {}
    }

    # 7) FOTOS: 1 a 4 fotos para cada perfil de forma consecutiva
    if ($fotos.Count -gt 0) {
        $cookieVal = $sess.Cookies.GetCookies($baseUrl)[$cookieName].Value
        if ($cookieVal) {
            $numFotos = Get-Random -Minimum 1 -Maximum 5
            for ($f = 0; $f -lt $numFotos; $f++) {
                $foto = $fotos | Get-Random
                $null = curl.exe -s -b "$cookieName=$cookieVal" `
                    -X POST "$baseUrl/api/v1/profiles/me/photos" `
                    -F "photo=@$foto;type=image/jpeg"
            }
        }
    }

    Write-Host " [OK]" -ForegroundColor Green
    if ($delayMs -gt 0) { Start-Sleep -Milliseconds $delayMs }
}

# ============================================================================
# FASE 2: Generar Interacciones MASIVAS (Likes, Matches, Favoritos y Visitas)
# ============================================================================
Write-Host "`n=== 2/2 Generando Interacciones de red para user1 ===" -ForegroundColor Green

function Login-Demo($userNum) {
    $em = "user$userNum@datingdemo.com"
    $b = @{ email = $em; password = "DemoPass123!" } | ConvertTo-Json
    $ip = "192.168.2.$($userNum % 250 + 1)"
    $headers = @{ "X-Forwarded-For" = $ip; "X-Real-IP" = $ip }
    
    $s = New-Object Microsoft.PowerShell.Commands.WebRequestSession
    for ($att = 1; $att -le 4; $att++) {
        try {
            $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post `
                -Body $b -ContentType "application/json" -Headers $headers -WebSession $s
            return $s
        } catch {
            Clear-RateLimits
            Start-Sleep -Milliseconds 200
        }
    }
    return $s
}

$user1Session = Login-Demo 1
$user1ProfileId = $profileIdMap[1]

if ($user1ProfileId) {
    # 1. Matches (User 2 a 15 hacen like mutuo con user1)
    for ($u = 2; $u -le 15; $u++) {
        $uSess = Login-Demo $u
        $uProfileId = $profileIdMap[$u]
        if ($uProfileId) {
            try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/likes/$user1ProfileId" -Method Post -WebSession $uSess } catch {}
            try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/likes/$uProfileId" -Method Post -WebSession $user1Session } catch {}
        }
    }
    Write-Host " -> 14 Matches creados para user1" -ForegroundColor Cyan

    # 2. Likes Recibidos pendientes (User 16 a 40 le dan like a user1)
    for ($u = 16; $u -le 40; $u++) {
        $uSess = Login-Demo $u
        try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/likes/$user1ProfileId" -Method Post -WebSession $uSess } catch {}
    }
    Write-Host " -> 25 Likes recibidos creados para user1" -ForegroundColor Cyan

    # 3. Likes Enviados (user1 le da like a User 41 a 55)
    for ($u = 41; $u -le 55; $u++) {
        $uProfileId = $profileIdMap[$u]
        if ($uProfileId) {
            try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/likes/$uProfileId" -Method Post -WebSession $user1Session } catch {}
        }
    }
    Write-Host " -> 15 Likes enviados creados por user1" -ForegroundColor Cyan

    # 4. Favoritos MUTUOS (User 56 a 65 y user1 se marcan mutuamente como favorito)
    for ($u = 56; $u -le 65; $u++) {
        $uSess = Login-Demo $u
        $uProfileId = $profileIdMap[$u]
        if ($uProfileId) {
            try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/favorites/$user1ProfileId" -Method Post -WebSession $uSess } catch {}
            try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/favorites/$uProfileId" -Method Post -WebSession $user1Session } catch {}
        }
    }
    Write-Host " -> 10 Favoritos MUTUOS creados" -ForegroundColor Cyan

    # 5. Favoritos Recibidos y Enviados cruzados
    for ($u = 66; $u -le 85; $u++) {
        $uSess = Login-Demo $u
        try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/favorites/$user1ProfileId" -Method Post -WebSession $uSess } catch {}
    }
    for ($u = 86; $u -le 100; $u++) {
        $uProfileId = $profileIdMap[$u]
        if ($uProfileId) {
            try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/favorites/$uProfileId" -Method Post -WebSession $user1Session } catch {}
        }
    }
    Write-Host " -> 20 Favoritos recibidos / 15 Favoritos enviados" -ForegroundColor Cyan

    # 6. Visitas (Usuarios 101 al 200 visitan a user1)
    for ($u = 101; $u -le 200; $u++) {
        if ($u % 25 -eq 0) { Clear-RateLimits }
        $uSess = Login-Demo $u
        try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/visits/$user1ProfileId" -Method Post -WebSession $uSess } catch {}
    }
    for ($u = 201; $u -le 230; $u++) {
        $uProfileId = $profileIdMap[$u]
        if ($uProfileId) {
            try { $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/visits/$uProfileId" -Method Post -WebSession $user1Session } catch {}
        }
    }
    Write-Host " -> 100 Visitas recibidas / 30 Visitas enviadas" -ForegroundColor Cyan
}

Write-Host "`n=== TODO LISTO: 500 usuarios generados con datos reales, fotos e interacciones masivas ===" -ForegroundColor Green
Write-Host "Inicia sesión en la aplicación con: user1@datingdemo.com / DemoPass123!" -ForegroundColor Yellow