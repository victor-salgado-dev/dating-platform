import styles from './page.module.css';

async function getBackendHealth() {
  // Al ser un Server Component dentro de Docker, debe llamar al contenedor backend directamente
  const apiUrl =
    process.env.INTERNAL_API_URL ?? 'http://backend:8080/api/v1';

  try {
    const res = await fetch(`${apiUrl}/health`, { cache: 'no-store' });
    const data = await res.json();
    return { ok: res.ok, data };
  } catch (err) {
    return { ok: false, data: null };
  }
}

export default async function HomePage() {
  const health = await getBackendHealth();

  return (
    <main className={styles.main}>
      <h1>Dating Platform</h1>
      <p>Fase 1: base del proyecto.</p>
      <p>
        Estado del backend:{' '}
        <strong>{health.ok ? 'operativo' : 'no disponible'}</strong>
      </p>
    </main>
  );
}