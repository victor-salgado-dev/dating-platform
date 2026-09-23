'use client';

// Este archivo ya no renderiza nada directamente: expone los datos de la
// cuenta (autenticación, foto, % de completado) como un hook, para que
// header-chrome.tsx pueda dibujar el avatar fusionado con la barra sin
// duplicar la lógica de peticiones.

import { useEffect, useState } from 'react';
import { usePathname } from 'next/navigation';
import { apiFetch, PublicProfile, ProfilePhoto } from '@/lib/api';

export interface AccountCompletion {
  percent: number;
  color: string;
}

export interface AccountData {
  authenticated: boolean | null;
  photo: string | null;
  completion: AccountCompletion;
}

const DEFAULT_COMPLETION: AccountCompletion = { percent: 0, color: '#ef4444' };

export function useAccountData(): AccountData {
  const pathname = usePathname();

  const [authenticated, setAuthenticated] = useState<boolean | null>(null);
  const [photo, setPhoto] = useState<string | null>(null);
  const [completion, setCompletion] = useState<AccountCompletion>(DEFAULT_COMPLETION);

  useEffect(() => {
    let active = true;

    setAuthenticated(null);
    apiFetch('/auth/me')
      .then(() => {
        if (!active) return;
        setAuthenticated(true);

        Promise.all([
          apiFetch<PublicProfile>('/profiles/me').catch(() => null),
          apiFetch<ProfilePhoto[]>('/profiles/me/photos').catch(() => []),
        ]).then(([profile, photos]) => {
          if (!active) return;

          if (photos && photos.length > 0) {
            setPhoto(photos[0].url ?? `/api/v1/profiles/me/photos/${photos[0].id}/file`);
          }

          let score = 0;
          const maxScore = 12; // Criterios totales

          if (profile) {
            score += 4; // Datos base (nombre, genero, fecha nac, pais) siempre existen si hay perfil
            if (profile.region) score++;
            if (profile.relationship_goals && profile.relationship_goals.length > 0) score++;
            if (profile.has_children !== null) score++;
            if (profile.bio) score++;
            if (profile.wants_children !== null) score++;
            if (profile.nationality) score++;
          }
          if (photos && photos.length > 0) score += 2; // Extra por foto

          const percent = Math.round((score / maxScore) * 100);

          let color = '#ef4444'; // Rojo por defecto
          if (percent >= 50 && percent < 80) color = '#eab308'; // Amarillo
          if (percent >= 80) color = '#22c55e'; // Verde

          setCompletion({ percent, color });
        });
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

  return { authenticated, photo, completion };
}

export async function logoutAndRedirect(router: { push: (href: string) => void }) {
  try {
    await apiFetch<void>('/auth/logout', { method: 'POST' });
  } catch {
    // Si falla, igualmente cerramos sesión en el cliente
  }
  window.dispatchEvent(new Event('auth-change'));
  router.push('/login');
}
