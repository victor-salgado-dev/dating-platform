'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

import { apiFetch } from '@/lib/api';

export default function AccountNav() {
  const pathname = usePathname();
  const [authenticated, setAuthenticated] = useState<boolean | null>(null);

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

  if (authenticated === null) return null;

  return authenticated ? (
    <Link href="/account">Cuenta</Link>
  ) : (
    <Link href="/login">Iniciar sesión</Link>
  );
}