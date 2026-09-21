# -*- coding: utf-8 -*-
import re
import unicodedata

def slugify(label):
    # Quita el contenido entre paréntesis para la clave (pero no para el label)
    base = re.sub(r'\(.*?\)', '', label)
    nfkd = unicodedata.normalize('NFKD', base)
    ascii_str = nfkd.encode('ascii', 'ignore').decode('ascii')
    ascii_str = ascii_str.lower()
    ascii_str = re.sub(r'[^a-z0-9]+', '_', ascii_str)
    return ascii_str.strip('_')

# =====================================================================
# GRUPO 1: intereses con intensidad 1-5 (has_level = true)
# =====================================================================
LEVELED = [
    ("sport_activity", "Deporte y actividad física", [
        "Deporte / actividad física", "Fitness / gimnasio", "Running", "Ciclismo",
        "Senderismo / montaña", "Deportes de equipo", "Deportes acuáticos",
        "Deportes de invierno", "Deportes de combate", "Atletismo", "Baile",
        "Yoga / pilates", "Equitación", "Escalada",
    ]),
    ("creativity_manual", "Creatividad y actividades manuales", [
        "Bricolaje / DIY", "Carpintería", "Restauración", "Manualidades", "Artesanía",
        "Costura", "Tejido / crochet", "Jardinería", "Plantas / flores",
        "Decoración / interiorismo", "Fotografía", "Dibujo / pintura", "Escritura",
        "Cocina", "Repostería",
    ]),
    ("culture_intellectual", "Cultura e intereses intelectuales", [
        "Lectura", "Ciencia", "Tecnología / informática", "Programación",
        "Arte y cultura", "Historia", "Filosofía", "Psicología",
        "Aprendizaje / formación", "Música", "Cine / series", "Teatro",
    ]),
    ("leisure_entertainment", "Ocio y entretenimiento", [
        "Videojuegos", "Juegos de mesa", "Juegos de cartas", "Viajar", "Naturaleza",
        "Camping", "Excursiones", "Compras", "Salir a comer", "Cafés",
        "Vida nocturna / clubs", "Fiestas", "Conciertos / festivales", "Socializar",
        "Conocer gente nueva", "Actividades con amigos",
    ]),
    ("lifestyle_other", "Estilo de vida y otros", [
        "Animales", "Coches / motor", "Moda", "Belleza / cosmética", "Bienestar",
        "Meditación", "Voluntariado",
    ]),
]

# =====================================================================
# GRUPO 2: hobbies/intereses concretos (has_level = false)
# Duplicados literales dentro de este mismo grupo, resueltos quedándose
# en una sola categoría (la más específica):
#   Cerámica  -> se queda en diy_crafts, se quita de art_creativity
#   Camping   -> se queda en nature_animals, se quita de travel
#   Pesca     -> se queda en nature_animals, se quita de sports_specific
#   Karting   -> se queda en motor, se quita de sports_specific
#   Anime     -> se queda en gaming_geek, se quita de film_entertainment
# =====================================================================
CONCRETE = [
    ("art_creativity", "Arte y creatividad", [
        "Dibujo", "Pintura", "Acuarela", "Óleo", "Ilustración", "Escultura",
        "Fotografía", "Fotografía de naturaleza", "Fotografía urbana", "Escritura",
        "Poesía", "Caligrafía", "Origami", "Modelismo", "Maquetas", "Scrapbooking",
        "Coleccionismo",
    ]),
    ("diy_crafts", "Manualidades / DIY", [
        "Bricolaje", "Carpintería", "Restauración de muebles", "Electrónica DIY",
        "Impresión 3D", "Cerámica", "Costura", "Bordado", "Punto", "Crochet",
        "Cuero", "Fabricación de velas", "Fabricación de jabón",
        "Reparaciones domésticas",
    ]),
    ("music", "Música", [
        "Cantar", "Karaoke", "Guitarra", "Guitarra eléctrica", "Bajo", "Piano",
        "Teclado", "Batería", "Violín", "Violonchelo", "Saxofón", "Flauta",
        "Otros instrumentos", "Composición musical", "Producción musical", "DJ",
    ]),
    ("music_genres", "Géneros musicales", [
        "Pop", "Rock", "Hard rock", "Metal", "Punk", "Indie", "Alternative",
        "Grunge", "Electrónica", "Techno", "House", "Trance", "EDM", "Hip hop",
        "Rap", "R&B", "Soul", "Funk", "Reggae", "Ska", "Jazz", "Blues",
        "Música clásica", "Country", "Folk", "Flamenco", "Música latina",
        "Reggaeton", "Salsa", "Bachata", "Tango", "K-pop", "J-pop", "Gospel",
        "Bandas sonoras",
    ]),
    ("gaming_geek", "Gaming / geek", [
        "Videojuegos", "PC gaming", "PlayStation", "Xbox", "Nintendo",
        "Juegos móviles", "Juegos retro", "RPG", "MMORPG", "Juegos de estrategia",
        "Juegos de simulación", "Juegos de cartas", "Juegos de mesa",
        "Dungeons & Dragons", "Anime", "Manga", "Cosplay", "Cómics",
        "Ciencia ficción", "Fantasía", "Tecnología", "Programación", "Robótica",
        "Inteligencia artificial",
    ]),
    ("sports_specific", "Deportes concretos", [
        "Fútbol", "Baloncesto", "Tenis", "Pádel", "Voleibol", "Bádminton", "Golf",
        "Béisbol", "Rugby", "Hockey", "Boxeo", "MMA", "Judo", "Karate",
        "Taekwondo", "Esquí", "Snowboard", "Surf", "Windsurf", "Kitesurf",
        "Buceo", "Kayak", "Piragüismo", "Natación", "Triatlón", "Gimnasia",
        "Patinaje", "Escalada", "Equitación", "Mountain bike", "Motociclismo",
        "Automovilismo",
    ]),
    ("motor", "Motor", [
        "Coches", "Motocicletas", "Coches clásicos", "Motos clásicas", "Motorsport",
        "Mecánica", "Tuning", "Restauración de vehículos", "Conducción", "Karting",
    ]),
    ("nature_animals", "Naturaleza y animales", [
        "Perros", "Gatos", "Caballos", "Pájaros", "Peces / acuarios", "Reptiles",
        "Roedores", "Animales de granja", "Apicultura", "Pesca",
        "Observación de aves", "Camping", "Senderismo", "Montañismo",
        "Jardinería", "Plantas", "Flores", "Huerto",
    ]),
    ("travel", "Viajes", [
        "Viajes", "Escapadas urbanas", "Viajes de naturaleza", "Viajes de playa",
        "Viajes de montaña", "Road trips", "Mochilero / backpacking",
        "Camper / autocaravana", "Cruceros", "Turismo cultural",
        "Turismo gastronómico", "Viajes de aventura", "Viajes de lujo",
    ]),
    ("gastronomy", "Gastronomía", [
        "Cocina italiana", "Cocina española", "Cocina francesa", "Cocina griega",
        "Cocina portuguesa", "Cocina mexicana", "Cocina argentina",
        "Cocina brasileña", "Cocina peruana", "Cocina colombiana",
        "Cocina japonesa", "Cocina china", "Cocina coreana", "Cocina tailandesa",
        "Cocina vietnamita", "Cocina india", "Cocina turca", "Cocina árabe",
        "Cocina libanesa", "Cocina marroquí", "Cocina africana", "Sushi",
        "Barbacoa", "Street food", "Comida picante", "Marisco", "Pescado",
        "Carne", "Dulces", "Postres", "Comida saludable", "Comida vegetariana",
        "Comida vegana",
    ]),
    ("film_entertainment", "Cine y entretenimiento", [
        "Acción", "Aventuras", "Comedia", "Drama", "Romance", "Thriller",
        "Terror", "Ciencia ficción", "Fantasía", "Misterio", "Crimen",
        "Documentales", "Animación", "Historia", "Guerra", "Musicales",
        "Western", "Reality",
    ]),
    ("books", "Libros", [
        "Novelas", "Romance", "Thriller", "Misterio", "Fantasía",
        "Ciencia ficción", "Terror", "Historia", "Biografías", "Psicología",
        "Ciencia", "Filosofía", "Negocios", "Desarrollo personal", "Poesía",
        "Cómics", "Manga",
    ]),
]

def build_rows():
    rows = []  # (key, category, label, has_level, sort_order)
    seen_keys = {}

    def add(category, label, has_level, sort_order, key_suffix=None):
        key = slugify(label)
        if key_suffix:
            key = f"{key}_{key_suffix}"
        if key in seen_keys:
            raise SystemExit(f"CLAVE DUPLICADA: {key!r} ya usada por {seen_keys[key]!r}, ahora para {(category, label)!r}")
        seen_keys[key] = (category, label)
        rows.append((key, category, label, has_level, sort_order))

    leveled_labels = {label for _cat, _es, items in LEVELED for label in items}

    for category, _label_es, items in LEVELED:
        for i, label in enumerate(items, start=1):
            add(category, label, True, i)

    # Si una etiqueta EXACTA ya existe en el grupo con intensidad, no tiene
    # sentido repetirla sin intensidad en el grupo concreto: la intensidad
    # ya cubre "lo tengo / no lo tengo" y más. Se omite del catálogo de
    # concretos (ej: "Escalada", "Camping", "Ciencia"...).
    skipped_as_leveled = []
    skipped_as_dup_concrete = []
    seen_concrete_labels = set()

    # Estas etiquetas se repiten entre "Cine y entretenimiento" y "Libros"
    # con el mismo texto pero significan cosas distintas (género de
    # película vs género de libro), así que se desambiguan con un sufijo
    # en la CLAVE en vez de fundirse en una sola fila.
    FILM_BOOK_AMBIGUOUS = {"Romance", "Thriller", "Terror", "Ciencia ficción", "Fantasía", "Misterio", "Historia"}

    for category, _label_es, items in CONCRETE:
        i = 0
        for label in items:
            if label in leveled_labels:
                skipped_as_leveled.append((category, label))
                continue

            suffix = None
            if label in FILM_BOOK_AMBIGUOUS and category in ("film_entertainment", "books"):
                suffix = "film" if category == "film_entertainment" else "book"

            # El chequeo de "ya visto" solo aplica a las que NO se van a
            # desambiguar por sufijo: esas dos sí pueden coexistir.
            if suffix is None and label in seen_concrete_labels:
                skipped_as_dup_concrete.append((category, label))
                continue

            seen_concrete_labels.add(label)
            i += 1
            add(category, label, False, i, key_suffix=suffix)

    if skipped_as_leveled or skipped_as_dup_concrete:
        import sys
        if skipped_as_leveled:
            print("-- Omitidas del grupo concreto por existir ya con intensidad:", file=sys.stderr)
            for cat, lbl in skipped_as_leveled:
                print(f"--   {cat}: {lbl}", file=sys.stderr)
        if skipped_as_dup_concrete:
            print("-- Omitidas del grupo concreto por estar repetidas en otra categoría concreta:", file=sys.stderr)
            for cat, lbl in skipped_as_dup_concrete:
                print(f"--   {cat}: {lbl}", file=sys.stderr)

    return rows

def sql_escape(s):
    return s.replace("'", "''")

def main():
    rows = build_rows()
    print(f"-- Total de filas generadas: {len(rows)}")
    print("INSERT INTO interests (key, category, label, has_level, sort_order) VALUES")
    lines = []
    for key, category, label, has_level, sort_order in rows:
        lines.append(
            f"    ('{key}', '{category}', '{sql_escape(label)}', {'true' if has_level else 'false'}, {sort_order})"
        )
    print(",\n".join(lines) + ";")

if __name__ == "__main__":
    main()
