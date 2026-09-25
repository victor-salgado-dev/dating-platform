# seed_profiles.ps1
#
# Puebla la base de datos con 500 usuarios demo que cubren TODOS los campos
# introducidos por las migraciones 000012–000018, a través de la API HTTP
# pública (no INSERTs directos), de forma que se ejecute también la
# validación de dominio, la regla de datos faltantes y los handlers.
#
#   - 000012: has_children, wants_children, height, weight, body_type,
#             ethnicity, appearance_rating, hair_color, eye_color, body_art,
#             smoking_habit, drinking_habit, relocation_willingness,
#             marital_status, children_count, youngest/oldest_child_age,
#             occupation, employment_status, income_level, living_situation,
#             nationality, education_level, english_ability, religion,
#             religious_values, star_sign
#   - 000013: future_vision, sports, likes_pets, pets_owned, favorite_season,
#             ideal_vacation_style, vacation_activities, profile_quote,
#             dream_wish, profile_partner_preferences, personality_statements,
#             profile_personality_answers
#   - 000014: profile_languages (código + nivel)
#   - 000015: relationship_goals (multi-select)
#   - 000016: interests (catálogo) + profile_interests
#   - 000004: profile_photos (multipart, rotando las fotos locales)
#
#   - 000017: profile_visits NO se puebla aquí: se crea al visitar perfiles
#             desde la app (no hay endpoint para forzar visitas históricas).
#   - 000018: solo crea índices, no requiere datos.
#
# Uso:
#   pwsh -File .\seed_profiles.ps1
#   (o ejecutarlo directamente desde una consola de PowerShell 5.1/7+)
#
# Requisitos:
#   - Backend corriendo en $baseUrl con las migraciones aplicadas.
#   - curl.exe disponible en el PATH (viene con Windows 10+).
#
# ⚠️ Rate limit: /auth/register está limitado por IP (AuthRateLimit). Con
#     500 registros seguidos puedes recibir 429. Si eso pasa: sube
#     AuthRateLimit en el .env del backend, o sube $delayMs. El script es
#     register-or-login, así que puedes reejecutarlo sin duplicar usuarios.
#
# ⚠️ PowerShell 5.1 aplana arrays de un solo elemento en ConvertTo-Json.
#     Este script siempre envía arrays de >= 2 elementos en los campos
#     multi-valor, tanto para evitarlo como para ejercitar el multi-select.

# --- Configuración ----------------------------------------------------------
$baseUrl    = "http://localhost"
$totalUsers = 500
$delayMs    = 0       # p.ej. 50 si pegas con el rate limit
$cookieName = "session_id"
$fotosDir   = "C:\Users\Victor\Downloads\fotos\Nueva carpeta (4)"

# --- Fotos disponibles ------------------------------------------------------
$fotos = @()
if (Test-Path $fotosDir) {
    $fotos = Get-ChildItem -Path $fotosDir -Filter "*.jpg" | Select-Object -ExpandProperty FullName
}
if ($fotos.Count -eq 0) {
    Write-Host "Aviso: no se encontraron .jpg en $fotosDir — se crearán perfiles sin foto." -ForegroundColor Yellow
}

# --- Datos base: nombres, géneros, bios, regiones ---------------------------
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

$regiones = @("Madrid","Barcelona","Valencia","Sevilla","Bilbao","Málaga","Zaragoza","Murcia","Alicante","Granada")

# --- Enums alineados con los CHECK de las migraciones 000012 y 000013 ------
$hasChildrenOptions     = @("yes","no","prefer_not_to_say")
$wantsChildrenOptions   = @("yes","no","not_sure")
$relationshipGoalOptions= @("casual","long_term","friendship","marriage","not_sure")
$bodyTypes              = @("petite","slim","athletic","average","few_extra_pounds","full_figured","large_and_lovely")
$ethnicities            = @("arab","asian","black","caucasian","hispanic","indian","mixed","pacific_islander","other")
$appearanceRatings      = @("below_average","average","attractive","very_attractive")
$hairColors             = @("bald","black","blonde","brown","grey","light_brown","red","changes_frequently","other")
$eyeColors              = @("black","blue","brown","green","grey","hazel","other")
$bodyArtOptions         = @("branding","earrings","piercing","tattoo","other","none")
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
$petsOwnedOptions       = @("none","cat","dog","horse","other")
$seasons                = @("spring","summer","autumn","winter")
$vacationStyles         = @("small_charming_hotel","luxury_hotel","cruise_ship","club_hotel","rental_apartment","countryside_house","camping_rv","staying_home","at_friends")
$vacationActivitiesOpts = @("cafes_shopping_nightlife","mix_relaxation_activities","lazing_and_relaxing","beach_holiday","sightseeing_cities","lots_of_sports")

$occupations = @(
    "administrative","advertising","artistic","construction","domestic_helper",
    "education","entertainment","executive","farming","finance",
    "fire_law_enforcement","hair_dresser","it_communications","laborer","legal",
    "medical","military","nanny","none","non_profit","political","retail",
    "retired","sales","self_employed","sports","student","technical",
    "transportation","travel","unemployed","other"
)

$quotes = @(
    "La vida es eso que pasa mientras haces otros planes.",
    "Colecciona momentos, no cosas.",
    "Hazlo con pasión o no lo hagas.",
    "Cada día es una nueva oportunidad.",
    "Ríe fuerte, ama sin miedo.",
    "Viaja mientras puedas, ama mientras quieras.",
    "El mejor proyecto en el que trabajar: tú.",
    "Menos scroll, más vida real."
)

$dreams = @(
    "Recorrer Japón en tren con una mochila y una cámara.",
    "Aprender a tocar el piano y dar un pequeño concierto para amigos.",
    "Vivir un año en otra ciudad sin ningún plan prefijado.",
    "Abrir un pequeño café-librería de barrio.",
    "Ver la aurora boreal desde una cabaña en Noruega.",
    "Sailing por el Mediterráneo durante un verano.",
    "Escribir una novela (aunque no la publique).",
    "Hacer un road trip por la costa oeste de EE.UU."
)

# --- Idiomas permitidos (CHECK de 000014) ----------------------------------
$languageCodes = @("es","en","fr","de","it","pt","ja","zh_cmn","ar","ru")

# --- Claves de intereses con has_level = true (000016) ---------------------
$leveledInterestKeys = @(
    # sport_activity
    "deporte_actividad_fisica","fitness_gimnasio","running","ciclismo","senderismo_montana",
    "deportes_de_equipo","deportes_acuaticos","deportes_de_invierno","deportes_de_combate",
    "atletismo","baile","yoga_pilates","equitacion","escalada",
    # creativity_manual
    "bricolaje_diy","carpinteria","restauracion","manualidades","artesania","costura",
    "tejido_crochet","jardineria","plantas_flores","decoracion_interiorismo","fotografia",
    "dibujo_pintura","escritura","cocina","reposteria",
    # culture_intellectual
    "lectura","ciencia","tecnologia_informatica","programacion","arte_y_cultura","historia",
    "filosofia","psicologia","aprendizaje_formacion","musica","cine_series","teatro",
    # leisure_entertainment
    "videojuegos","juegos_de_mesa","juegos_de_cartas","viajar","naturaleza","camping",
    "excursiones","compras","salir_a_comer","cafes","vida_nocturna_clubs","fiestas",
    "conciertos_festivales","socializar","conocer_gente_nueva","actividades_con_amigos",
    # lifestyle_other
    "animales","coches_motor","moda","belleza_cosmetica","bienestar","meditacion","voluntariado"
)

# --- Claves de intereses con has_level = false (000016) --------------------
$simpleInterestKeys = @(
    # art_creativity
    "dibujo","pintura","acuarela","oleo","ilustracion","escultura",
    "fotografia_de_naturaleza","fotografia_urbana","poesia","caligrafia","origami","modelismo",
    "maquetas","scrapbooking","coleccionismo",
    # diy_crafts
    "bricolaje","restauracion_de_muebles","electronica_diy","impresion_3d","ceramica","bordado",
    "punto","crochet","cuero","fabricacion_de_velas","fabricacion_de_jabon","reparaciones_domesticas",
    # music
    "cantar","karaoke","guitarra","guitarra_electrica","bajo","piano","teclado","bateria",
    "violin","violonchelo","saxofon","flauta","otros_instrumentos","composicion_musical",
    "produccion_musical","dj",
    # music_genres
    "pop","rock","hard_rock","metal","punk","indie","alternative","grunge","electronica","techno",
    "house","trance","edm","hip_hop","rap","r_b","soul","funk","reggae","ska","jazz","blues",
    "musica_clasica","country","folk","flamenco","musica_latina","reggaeton","salsa","bachata",
    "tango","k_pop","j_pop","gospel","bandas_sonoras",
    # gaming_geek
    "pc_gaming","playstation","xbox","nintendo","juegos_moviles","juegos_retro","rpg","mmorpg",
    "juegos_de_estrategia","juegos_de_simulacion","dungeons_dragons","anime","manga","cosplay",
    "comics","ciencia_ficcion","fantasia","tecnologia","robotica","inteligencia_artificial",
    # sports_specific
    "futbol","baloncesto","tenis","padel","voleibol","badminton","golf","beisbol","rugby","hockey",
    "boxeo","mma","judo","karate","taekwondo","esqui","snowboard","surf","windsurf","kitesurf",
    "buceo","kayak","piraguismo","natacion","triatlon","gimnasia","patinaje","mountain_bike",
    "motociclismo","automovilismo",
    # motor
    "coches","motocicletas","coches_clasicos","motos_clasicas","motorsport","mecanica","tuning",
    "restauracion_de_vehiculos","conduccion","karting",
    # nature_animals
    "perros","gatos","caballos","pajaros","peces_acuarios","reptiles","roedores","animales_de_granja",
    "apicultura","pesca","observacion_de_aves","senderismo","montanismo","plantas","flores","huerto",
    # travel
    "viajes","escapadas_urbanas","viajes_de_naturaleza","viajes_de_playa","viajes_de_montana",
    "road_trips","mochilero_backpacking","camper_autocaravana","cruceros","turismo_cultural",
    "turismo_gastronomico","viajes_de_aventura","viajes_de_lujo",
    # gastronomy
    "cocina_italiana","cocina_espanola","cocina_francesa","cocina_griega","cocina_portuguesa",
    "cocina_mexicana","cocina_argentina","cocina_brasilena","cocina_peruana","cocina_colombiana",
    "cocina_japonesa","cocina_china","cocina_coreana","cocina_tailandesa","cocina_vietnamita",
    "cocina_india","cocina_turca","cocina_arabe","cocina_libanesa","cocina_marroqui","cocina_africana",
    "sushi","barbacoa","street_food","comida_picante","marisco","pescado","carne","dulces","postres",
    "comida_saludable","comida_vegetariana","comida_vegana",
    # film_entertainment
    "accion","aventuras","comedia","drama","romance_film","thriller_film","terror_film",
    "ciencia_ficcion_film","fantasia_film","misterio_film","crimen","documentales","animacion",
    "guerra","musicales","western","reality",
    # books
    "novelas","romance_book","thriller_book","misterio_book","fantasia_book","ciencia_ficcion_book",
    "terror_book","biografias","negocios","desarrollo_personal"
)

# --- Afirmaciones de personalidad (000013) ---------------------------------
$personalityKeys = @(
    "extra_reserved_calm","extra_funny_laughs","extra_center_of_party","extra_enjoys_alone_time",
    "emo_sensitive_vulnerable","emo_moody","emo_self_confident","emo_hard_to_rattle",
    "cons_chaotic","cons_no_planning","cons_goal_oriented","cons_very_tidy",
    "agree_distrustful_at_first","agree_helpful_caring","agree_hard_to_get_along","agree_believes_in_good",
    "open_original_new_ideas","open_cautious_with_new","open_interested_arts","open_traditions_matter"
)

# --- Partner preferences: valores permitidos (000013) ----------------------
$desiredTraitsOptions = @(
    "humorous","self_confident","loving","kind_hearted","intelligent","faithful","honest",
    "ambitious","family_oriented","adventurous","romantic","patient","easy_going",
    "financially_stable","other"
)
$partnerMayHaveChildrenOptions    = @("yes","no","doesnt_matter")
$partnerReligionOptions           = $religions + @("doesnt_matter")
$firstMeetingOptions              = @("doesnt_matter","public_place","my_city","their_city","video_call_first")
$desiredLivingPlaceOptions        = @("big_city","medium_city","small_town","countryside","abroad")

# ============================================================================
# Bucle principal
# ============================================================================
Write-Host "=== Creando $totalUsers perfiles demo (cobertura completa de migraciones) ===" -ForegroundColor Green

for ($i = 0; $i -lt $totalUsers; $i++) {
    $num       = $i + 1
    $email     = "user$num@datingdemo.com"
    $pass      = "DemoPass123!"
    $nombre    = $nombres[$i % $nombres.Count]
    $genero    = $generos[$i % $generos.Count]
    $bio       = $bios[$i % $bios.Count]

    # birth_date: todos adultos (año 1985..2002 -> ~23..40 años)
    $year      = 1985 + ($i % 18)
    $month     = (($i % 12) + 1).ToString("00")
    $day       = (($i % 27) + 1).ToString("00")
    $birthDate = "$year-$month-$day"

    Write-Host ("[{0}/{1}] {2} ({3})..." -f $num, $totalUsers, $nombre, $email) -NoNewline

    # ------------------------------------------------------------------
    # 1) Auth: register-or-login (cookie HttpOnly en la WebSession)
    # ------------------------------------------------------------------
    $bodyReg = @{
        email          = $email
        password       = $pass
        accepted_terms = $true
    } | ConvertTo-Json

    try {
        $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" -Method Post `
            -Body $bodyReg -ContentType "application/json" -SessionVariable sess
    } catch {
        $bodyLogin = @{ email = $email; password = $pass } | ConvertTo-Json
        $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post `
            -Body $bodyLogin -ContentType "application/json" -SessionVariable sess
    }

    # ------------------------------------------------------------------
    # 2) Perfil (POST /profiles/me) con TODOS los campos opcionales
    #    Los arrays se envían siempre con 2 valores distintos (offset=1)
    #    para evitar el aplanado de PS 5.1 y probar el multi-select.
    # ------------------------------------------------------------------
    $goals      = @($relationshipGoalOptions[$i % $relationshipGoalOptions.Count],
                    $relationshipGoalOptions[($i + 1) % $relationshipGoalOptions.Count])
    $bodyArt    = @($bodyArtOptions[$i % $bodyArtOptions.Count],
                    $bodyArtOptions[($i + 1) % $bodyArtOptions.Count])
    $relocArr   = @($relocOptions[$i % $relocOptions.Count],
                    $relocOptions[($i + 1) % $relocOptions.Count])
    $futVis     = @($futureVisionOptions[$i % $futureVisionOptions.Count],
                    $futureVisionOptions[($i + 1) % $futureVisionOptions.Count])
    $sportsArr  = @($sportsOptions[$i % $sportsOptions.Count],
                    $sportsOptions[($i + 1) % $sportsOptions.Count])
    $petsArr    = @($petsOwnedOptions[$i % $petsOwnedOptions.Count],
                    $petsOwnedOptions[($i + 1) % $petsOwnedOptions.Count])
    $vacStyle   = @($vacationStyles[$i % $vacationStyles.Count],
                    $vacationStyles[($i + 1) % $vacationStyles.Count])
    $vacAct     = @($vacationActivitiesOpts[$i % $vacationActivitiesOpts.Count],
                    $vacationActivitiesOpts[($i + 1) % $vacationActivitiesOpts.Count])

    $hasChildren = $hasChildrenOptions[$i % $hasChildrenOptions.Count]

    $profileBody = [ordered]@{
        display_name           = $nombre
        birth_date             = $birthDate
        gender                 = $genero
        country_code           = "ES"
        region                 = $regiones[$i % $regiones.Count]
        relationship_goals     = $goals
        has_children           = $hasChildren
        wants_children         = $wantsChildrenOptions[$i % $wantsChildrenOptions.Count]
        bio                    = $bio

        height                 = 150 + ($i % 55)     # 150..204 cm
        weight                 = 50  + ($i % 60)     # 50..109 kg
        body_type              = $bodyTypes[$i % $bodyTypes.Count]
        ethnicity              = $ethnicities[$i % $ethnicities.Count]
        appearance_rating      = $appearanceRatings[$i % $appearanceRatings.Count]
        hair_color             = $hairColors[$i % $hairColors.Count]
        eye_color              = $eyeColors[$i % $eyeColors.Count]
        body_art               = $bodyArt

        smoking_habit          = $smokeOptions[$i % $smokeOptions.Count]
        drinking_habit         = $drinkOptions[$i % $drinkOptions.Count]
        relocation_willingness = $relocArr
        marital_status         = $maritalStatuses[$i % $maritalStatuses.Count]
        occupation             = $occupations[$i % $occupations.Count]
        employment_status      = $employmentStatuses[$i % $employmentStatuses.Count]
        income_level           = $incomeLevels[$i % $incomeLevels.Count]
        living_situation       = $livingSituations[$i % $livingSituations.Count]

        nationality            = "ES"
        education_level        = $educationLevels[$i % $educationLevels.Count]
        english_ability        = $englishAbilities[$i % $englishAbilities.Count]
        religion               = $religions[$i % $religions.Count]
        religious_values       = $religiousValues[$i % $religiousValues.Count]
        star_sign              = $starSigns[$i % $starSigns.Count]

        future_vision          = $futVis
        sports                 = $sportsArr
        likes_pets             = $likesPetsOptions[$i % $likesPetsOptions.Count]
        pets_owned             = $petsArr
        favorite_season        = $seasons[$i % $seasons.Count]
        ideal_vacation_style   = $vacStyle
        vacation_activities    = $vacAct
        profile_quote          = $quotes[$i % $quotes.Count]
        dream_wish             = $dreams[$i % $dreams.Count]
    }

    # Regla de datos faltantes: si no tiene hijos, no enviamos sus edades.
    if ($hasChildren -eq 'yes') {
        $profileBody.children_count    = 1 + ($i % 3)
        $profileBody.youngest_child_age= ($i % 12)
        $profileBody.oldest_child_age  = ($i % 12) + 5
    }

    # Huecos deliberados para probar NULL vs valor:
    if (($i % 17) -eq 0) {                       # texto opcional vacío
        $profileBody.Remove('profile_quote')  | Out-Null
        $profileBody.Remove('dream_wish')     | Out-Null
    }
    if (($i % 19) -eq 0) {                       # sin bio
        $profileBody.Remove('bio') | Out-Null
    }

    $profileJson = $profileBody | ConvertTo-Json -Depth 10

    try {
        $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me" -Method Post `
            -Body $profileJson -ContentType "application/json" -WebSession $sess
    } catch {
        Write-Host " [perfil]" -NoNewline -ForegroundColor DarkYellow
    }

    # ------------------------------------------------------------------
    # 3) Idiomas (000014): PUT /profiles/me/languages/{code}  body {level}
    #    1 de cada 13 usuarios se queda sin idiomas (probamos "no relleno").
    # ------------------------------------------------------------------
    if (($i % 13) -ne 0) {
        for ($k = 0; $k -lt 3; $k++) {
            $code = $languageCodes[($i + $k * 7) % $languageCodes.Count]
            $lvl  = 1 + (($i + $k) % 5)
            $langJson = @{ level = $lvl } | ConvertTo-Json
            try {
                $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/languages/$code" -Method Put `
                    -Body $langJson -ContentType "application/json" -WebSession $sess
            } catch { Write-Host " [lang:$code]" -NoNewline -ForegroundColor DarkYellow }
        }
    }

    # ------------------------------------------------------------------
    # 4) Intereses (000016): PUT /profiles/me/interests/{key}
    #    - has_level=true  -> body {level: 1..5}
    #    - has_level=false -> body {level: null}
    #    Se rotan las claves por índice para cubrir TODO el catálogo entre
    #    los 500 usuarios con un nº razonable de llamadas por usuario.
    # ------------------------------------------------------------------
    for ($k = 0; $k -lt 4; $k++) {
        $key = $leveledInterestKeys[($i + $k * 13) % $leveledInterestKeys.Count]
        $lvl = 1 + (($i + $k) % 5)
        $intJson = @{ level = $lvl } | ConvertTo-Json
        try {
            $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/interests/$key" -Method Put `
                -Body $intJson -ContentType "application/json" -WebSession $sess
        } catch { Write-Host " [int:$key]" -NoNewline -ForegroundColor DarkYellow }
    }
    for ($k = 0; $k -lt 6; $k++) {
        $key = $simpleInterestKeys[($i + $k * 17) % $simpleInterestKeys.Count]
        $intJson = @{ level = $null } | ConvertTo-Json
        try {
            $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/interests/$key" -Method Put `
                -Body $intJson -ContentType "application/json" -WebSession $sess
        } catch { Write-Host " [int:$key]" -NoNewline -ForegroundColor DarkYellow }
    }

    # ------------------------------------------------------------------
    # 5) Personalidad (000013): PUT /profiles/me/personality/{key}  {score}
    #    Se contestan las 20 afirmaciones (score rotando 1..5).
    # ------------------------------------------------------------------
    for ($p = 0; $p -lt $personalityKeys.Count; $p++) {
        $pk    = $personalityKeys[$p]
        $score = 1 + (($i + $p) % 5)
        $persJson = @{ score = $score } | ConvertTo-Json
        try {
            $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/personality/$pk" -Method Put `
                -Body $persJson -ContentType "application/json" -WebSession $sess
        } catch { Write-Host " [pers:$pk]" -NoNewline -ForegroundColor DarkYellow }
    }

    # ------------------------------------------------------------------
    # 6) Preferencias de pareja (000013): PATCH /profiles/me/partner-preferences
    #    1 de cada 11 usuarios se queda sin prefs.
    # ------------------------------------------------------------------
    if (($i % 11) -ne 0) {
        $traitsArr = @($desiredTraitsOptions[$i % $desiredTraitsOptions.Count],
                       $desiredTraitsOptions[($i + 1) % $desiredTraitsOptions.Count])
        $livArr    = @($desiredLivingPlaceOptions[$i % $desiredLivingPlaceOptions.Count],
                       $desiredLivingPlaceOptions[($i + 1) % $desiredLivingPlaceOptions.Count])

        $ageMin = 18 + ($i % 12)
        $hMin   = 150 + ($i % 30)

        $partnerBody = [ordered]@{
            age_min    = $ageMin
            age_max    = $ageMin + 5
            height_min = $hMin
            height_max = $hMin + 20

            desired_traits = $traitsArr

            partner_may_have_children    = $partnerMayHaveChildrenOptions[$i % $partnerMayHaveChildrenOptions.Count]
            partner_religion_preference  = $partnerReligionOptions[$i % $partnerReligionOptions.Count]
            about_partner_text           = "Busco a alguien con quien compartir $($regiones[$i % $regiones.Count]) y buenas conversaciones."
            first_meeting_preference     = $firstMeetingOptions[$i % $firstMeetingOptions.Count]
            desired_living_place         = $livArr

            importance_shared_thoughts    = 1 + (($i + 0) % 5)
            importance_shared_hobbies     = 1 + (($i + 1) % 5)
            importance_intimacy           = 1 + (($i + 2) % 5)
            importance_romantic_love      = 1 + (($i + 3) % 5)
            importance_financial_security = 1 + (($i + 4) % 5)
            importance_fun                = 1 + (($i + 5) % 5)
            importance_shared_friends     = 1 + (($i + 6) % 5)
            importance_shared_humor       = 1 + (($i + 7) % 5)
            importance_personal_space     = 1 + (($i + 8) % 5)
            importance_independence       = 1 + (($i + 9) % 5)
        }

        $partnerJson = $partnerBody | ConvertTo-Json -Depth 10
        try {
            $null = Invoke-RestMethod -Uri "$baseUrl/api/v1/profiles/me/partner-preferences" -Method Patch `
                -Body $partnerJson -ContentType "application/json" -WebSession $sess
        } catch { Write-Host " [partner]" -NoNewline -ForegroundColor DarkYellow }
    }

    # ------------------------------------------------------------------
    # 7) Foto (000004): multipart POST /profiles/me/photos
    #    Vía curl.exe porque Invoke-RestMethod con multipart es engorroso.
    #    1 de cada 11 usuarios se queda sin foto.
    # ------------------------------------------------------------------
    if ($fotos.Count -gt 0 -and ($i % 11) -ne 0) {
        $foto = $fotos[$i % $fotos.Count]
        $cookieVal = $sess.Cookies.GetCookies($baseUrl)[$cookieName].Value
        if ($cookieVal) {
            $subida = curl.exe -s -b "$cookieName=$cookieVal" `
                -X POST "$baseUrl/api/v1/profiles/me/photos" `
                -F "photo=@$foto;type=image/jpeg"
            if ($LASTEXITCODE -ne 0) {
                Write-Host " [foto]" -NoNewline -ForegroundColor DarkYellow
            }
        }
    }

    Write-Host " [OK]" -ForegroundColor Green

    if ($delayMs -gt 0) { Start-Sleep -Milliseconds $delayMs }
}

Write-Host ""
Write-Host "=== Listo: $totalUsers usuarios demo creados ===" -ForegroundColor Cyan
Write-Host "Credenciales: user1@datingdemo.com ... user$totalUsers@datingdemo.com / DemoPass123!" -ForegroundColor Cyan
Write-Host "Recuerda: profile_visits (000017) se genera solo al visitar perfiles desde la UI." -ForegroundColor Cyan
