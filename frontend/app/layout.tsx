import type { Metadata } from 'next';
import Link from 'next/link';
import './globals.css';
import styles from './layout.module.css';

export const metadata: Metadata = {
  title: 'Dating Platform',
  description: 'Plataforma internacional de dating/relaciones',
};

// Nota: en esta fase la interfaz solo tiene un idioma cableado (es-ES
// por defecto en el HTML). La estructura de mensajes en
// frontend/messages/{es,en}.json ya existe para que una fase futura
// dedicada a i18n pueda enrutar por idioma sin rehacer el layout.
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="es">
      <body>
        <nav className={styles.nav}>
          <Link href="/" className={styles.brand}>
            Dating Platform
          </Link>
          <Link href="/discover">Descubrir</Link>
          <Link href="/favorites">Favoritos</Link>
          <Link href="/messages">Mensajes</Link>
          <Link href="/blocked">Bloqueados</Link>
        </nav>
        {children}
      </body>
    </html>
  );
}
