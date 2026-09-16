function DiscoverContent() {
  const searchParams = useSearchParams();
  const searchString = searchParams.toString(); // Lo pasamos a string para evitar re-renders innecesarios
  
  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [lastSearch, setLastSearch] = useState(searchString);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Si los filtros cambian, reseteamos a la página 1 de forma segura (sin doble petición)
  if (searchString !== lastSearch) {
    setLastSearch(searchString);
    setPage(1);
  }

  useEffect(() => {
    let isMounted = true;
    setLoading(true);
    setError(null);

    const params = new URLSearchParams(searchString);
    params.set('page', page.toString());
    params.set('page_size', '24');

    apiFetch<SearchResponse>(`/search/profiles?${params.toString()}`)
      .then((res) => {
        if (isMounted) setData(res);
      })
      .catch((err: unknown) => {
        if (!isMounted) return;
        if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver perfiles.');
        } else if (err instanceof ApiError && err.status === 429) {
          setError('Demasiadas peticiones. Espera un minuto e inténtalo de nuevo.');
        } else {
          setError('No se pudieron cargar los resultados.');
        }
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });

    return () => {
      isMounted = false; // Cleanup para que el modo estricto de React no duplique el set state
    };
  }, [page, searchString]);

  const profiles = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <>
      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>Cargando perfiles…</p>}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {profiles.length === 0 ? (
            <div className={styles.emptyState}>
              <p>No hay resultados que coincidan con tu búsqueda.</p>
              <Link href="/search" style={{ color: 'var(--primary)', textDecoration: 'underline', marginTop: '1rem', display: 'inline-block' }}>
                Cambiar filtros
              </Link>
            </div>
          ) : (
            <ul className={styles.grid}>
              {profiles.map((item) => (
                <li key={item.profile_id} className={styles.card}>
                  <Link href={`/profiles/${item.profile_id}`} className={styles.cardLink}>
                    <div className={styles.imageContainer}>
                      {item.has_photo ? (
                        <ProfilePhoto profileId={item.profile_id} name={item.display_name} />
                      ) : (
                        <div className={styles.photoPlaceholder}>Sin Foto</div>
                      )}
                    </div>

                    <div className={styles.cardInfo}>
                      <h3 className={styles.name}>
                        {item.display_name}
                        <span className={styles.age}> · {item.age}</span>
                      </h3>
                      <p className={styles.details}>
                        {[item.region, item.country_code].filter(Boolean).join(', ')}
                      </p>
                      {item.relationship_goal && (
                        <span className={styles.cardGoal}>{item.relationship_goal}</span>
                      )}
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          )}

          {totalPages > 1 && (
            <div className={styles.pagination}>
              <button
                type="button"
                className={styles.pageButton}
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                ← Anterior
              </button>
              <span className={styles.pageInfo}>
                Página {data?.page} de {totalPages} ({data?.total} perfiles)
              </span>
              <button
                type="button"
                className={styles.pageButton}
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Siguiente →
              </button>
            </div>
          )}
        </>
      )}
    </>
  );
}