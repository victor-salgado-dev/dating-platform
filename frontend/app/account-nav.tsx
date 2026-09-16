'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';

import { apiFetch } from '@/lib/api';
import styles from './layout.module.css';

export default function AccountNav() {
  const pathname = usePathname();
  const router = useRouter();
  const [authenticated, setAuthenticated] = useState<boolean | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    let active = true;

    setAuthenticated(null);
    apiFetch('/auth/me')
      .then(() => {
        if (active) setAuthenticated(true);
      })
      .catch(() => {
        if (active) setAuthenticated(false);
      });

    return () => {
      active = false;
    };
  }, [pathname]);

  useEffect(() => {
    function handleAuthChange() {
      setAuthenticated(false);
    }

    window.addEventListener('auth-change', handleAuthChange);
    return () => window.removeEventListener('auth-change', handleAuthChange);
  }, []);

  async function handleLogout() {
    try {
      await apiFetch<void>('/auth/logout', { method: 'POST' });
    } catch {
    }
    setAuthenticated(false);
    setMenuOpen(false);
    window.dispatchEvent(new Event('auth-change'));
    router.push('/login');
  }

  if (authenticated === null) return null;

  if (!authenticated) {
    return (
      <Link href="/login">Iniciar sesión</Link>
    );
  }

  return (
    <div className={styles.accountMenu}>
      <button
        type="button"
        className={styles.accountButton}
        onClick={() => setMenuOpen((open) => !open)}
        aria-expanded={menuOpen}
        aria-haspopup="menu"
      >
        Cuenta
      </button>
      {menuOpen && (
        <div className={styles.accountDropdown} role="menu">
          <Link href="/profile" role="menuitem" onClick={() => setMenuOpen(false)}>
            Mi perfil
          </Link>
          <Link href="/settings" role="menuitem" onClick={() => setMenuOpen(false)}>
            Ajustes
          </Link>
          <button type="button" role="menuitem" onClick={handleLogout}>
            Cerrar sesión
          </button>
        </div>
      )}
    </div>
  );
}