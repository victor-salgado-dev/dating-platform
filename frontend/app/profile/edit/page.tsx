'use client';

import { ChangeEvent, FormEvent, useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { apiFetch, ApiError, ProfilePhoto, PublicProfile } from '@/lib/api';
import styles from './page.module.css';

type FormState = {
  display_name: string;
  birth_date: string;
  gender: string;
  country_code: string;
  region: string;
  languages: string;
  relationship_goal: string;
  has_children: string;
  wants_children: string;
  bio: string;
  interests: string;
};

const emptyForm: FormState = {
  display_name: '', birth_date: '', gender: '', country_code: '', region: '',
  languages: '', relationship_goal: '', has_children: '', wants_children: '',
  bio: '', interests: '',
};

function listValue(value: string) {
  return value.split(',').map((item) => item.trim()).filter(Boolean);
}

function nullableString(value: string) {
  return value.trim() || null;
}

function nullableBoolean(value: string) {
  return value === '' ? null : value === 'true';
}

function formFromProfile(profile: PublicProfile): FormState {
  return {
    display_name: profile.display_name,
    birth_date: '',
    gender: profile.gender,
    country_code: profile.country_code,
    region: profile.region ?? '',
    languages: profile.languages?.join(', ') ?? '',
    relationship_goal: profile.relationship_goal ?? '',
    has_children: profile.has_children === null ? '' : String(profile.has_children),
    wants_children: profile.wants_children === null ? '' : String(profile.wants_children),
    bio: profile.bio ?? '',
    interests: profile.interests?.join(', ') ?? '',
  };
}

export default function EditProfilePage() {
  const router = useRouter();
  const [form, setForm] = useState<FormState>(emptyForm);
  const [photos, setPhotos] = useState<ProfilePhoto[]>([]);
  const [creating, setCreating] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    apiFetch<PublicProfile>('/profiles/me')
      .then(async (profile) => {
        setForm(formFromProfile(profile));
        try {
          setPhotos(await apiFetch<ProfilePhoto[]>('/profiles/me/photos'));
        } catch {
          setPhotos([]);
        }
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) {
          setCreating(true);
        } else if (err instanceof ApiError && err.status === 401) {
          router.push('/login');
        } else {
          setError('No se pudo cargar el perfil.');
        }
      })
      .finally(() => setLoading(false));
  }, [router]);

  function updateField(field: keyof FormState, value: string) {
    setForm((current) => ({ ...current, [field]: value }));
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setError(null);
    setSaved(false);

    const body = {
      display_name: form.display_name,
      gender: form.gender,
      country_code: form.country_code,
      region: nullableString(form.region),
      languages: listValue(form.languages),
      relationship_goal: nullableString(form.relationship_goal),
      has_children: nullableBoolean(form.has_children),
      wants_children: nullableBoolean(form.wants_children),
      bio: nullableString(form.bio),
      interests: listValue(form.interests),
      ...(creating ? { birth_date: form.birth_date } : {}),
    };

    try {
      const profile = await apiFetch<PublicProfile>('/profiles/me', {
        method: creating ? 'POST' : 'PATCH',
        body: JSON.stringify(body),
      });
      setForm(formFromProfile(profile));
      setCreating(false);
      setSaved(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo guardar el perfil.');
    } finally {
      setSaving(false);
    }
  }

  async function handleUpload(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;

    setUploading(true);
    setError(null);
    const data = new FormData();
    data.append('photo', file);
    try {
      const photo = await apiFetch<ProfilePhoto>('/profiles/me/photos', {
        method: 'POST',
        body: data,
      });
      setPhotos((current) => [...current, photo]);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo subir la foto.');
    } finally {
      setUploading(false);
    }
  }

  async function handleDeletePhoto(photoID: string) {
    setError(null);
    try {
      await apiFetch<void>(`/profiles/me/photos/${photoID}`, { method: 'DELETE' });
      setPhotos((current) => current.filter((photo) => photo.id !== photoID));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo eliminar la foto.');
    }
  }

  if (loading) {
    return <main className={styles.main}><p>Cargando…</p></main>;
  }

  return (
    <main className={styles.main}>
      <Link href="/profile">&larr; Volver a mi perfil</Link>
      <h1>{creating ? 'Crear perfil' : 'Modificar perfil'}</h1>

      <form onSubmit={handleSubmit} className={styles.form}>
        <label>Nombre visible<input value={form.display_name} onChange={(e) => updateField('display_name', e.target.value)} required /></label>
        {creating && <label>Fecha de nacimiento<input type="date" value={form.birth_date} onChange={(e) => updateField('birth_date', e.target.value)} required /></label>}
        <label>Género
          <select value={form.gender} onChange={(e) => updateField('gender', e.target.value)} required>
            <option value="">Selecciona una opción</option><option value="female">Mujer</option><option value="male">Hombre</option><option value="non_binary">No binario</option><option value="other">Otro</option>
          </select>
        </label>
        <label>País (código de dos letras)<input value={form.country_code} onChange={(e) => updateField('country_code', e.target.value.toUpperCase())} maxLength={2} required /></label>
        <label>Región<input value={form.region} onChange={(e) => updateField('region', e.target.value)} /></label>
        <label>Idiomas <span className={styles.hint}>separados por comas</span><input value={form.languages} onChange={(e) => updateField('languages', e.target.value)} /></label>
        <label>Objetivo de relación
          <select value={form.relationship_goal} onChange={(e) => updateField('relationship_goal', e.target.value)}><option value="">Prefiero no decirlo</option><option value="casual">Algo casual</option><option value="long_term">Relación estable</option><option value="friendship">Amistad</option><option value="marriage">Matrimonio</option><option value="not_sure">No lo sé todavía</option></select>
        </label>
        <label>¿Tienes hijos?
          <select value={form.has_children} onChange={(e) => updateField('has_children', e.target.value)}><option value="">Prefiero no decirlo</option><option value="true">Sí</option><option value="false">No</option></select>
        </label>
        <label>¿Quieres tener hijos?
          <select value={form.wants_children} onChange={(e) => updateField('wants_children', e.target.value)}><option value="">Prefiero no decirlo</option><option value="true">Sí</option><option value="false">No</option></select>
        </label>
        <label>Biografía<textarea value={form.bio} onChange={(e) => updateField('bio', e.target.value)} maxLength={1000} rows={5} /></label>
        <label>Intereses <span className={styles.hint}>separados por comas</span><input value={form.interests} onChange={(e) => updateField('interests', e.target.value)} /></label>
        {error && <p className={styles.error}>{error}</p>}
        {saved && <p className={styles.success}>Perfil guardado.</p>}
        <button type="submit" disabled={saving}>{saving ? 'Guardando…' : 'Guardar perfil'}</button>
      </form>

      <section className={styles.photos}>
        <h2>Fotos</h2>
        {photos.length > 0 && <div className={styles.photoGrid}>{photos.map((photo) => <div key={photo.id} className={styles.photo}><img src={photo.url} alt="" /><button type="button" onClick={() => handleDeletePhoto(photo.id)}>Eliminar</button></div>)}</div>}
        {!creating && <label className={styles.upload}>Añadir foto<input type="file" accept="image/jpeg,image/png,image/webp" onChange={handleUpload} disabled={uploading} /></label>}
        {creating && <p className={styles.hint}>Guarda el perfil antes de añadir fotos.</p>}
      </section>
    </main>
  );
}