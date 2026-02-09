const API_BASE = import.meta.env.VITE_API_BASE || 'https://lucky-money-2026-production.up.railway.app'

let adminSession = localStorage.getItem('admin_session') || ''

export function setSession(s) {
  adminSession = s
  localStorage.setItem('admin_session', s)
}

function headers(isAdmin = false) {
  const h = { 'Content-Type': 'application/json' }
  if (isAdmin && adminSession) {
    h['X-Admin-Session'] = adminSession 
  }
  return h
}

export async function adminLogin(user, pass) {
  const res = await fetch(`${API_BASE}/admin/login`, {
    method: 'POST',
    headers: headers(),
    body: JSON.stringify({ user, pass }),
  })
  if (!res.ok) throw new Error('login failed')
  return res.json()
}

export async function createPool(items) {
  const res = await fetch(`${API_BASE}/admin/pool`, {
    method: 'POST',
    headers: headers(true),
    body: JSON.stringify({ items }),
  })
  if (!res.ok) throw new Error('save pool failed')
  return res.json()
}

export async function fetchCurrentPool() {
  const res = await fetch(`${API_BASE}/admin/pool`, {
    headers: headers(true),
  })
  if (!res.ok) throw new Error('fetch pool failed')
  return res.json()
}

export async function fetchClaimHistory() {
  const res = await fetch(`${API_BASE}/admin/claims`, {
    headers: headers(true),
  })
  if (!res.ok) throw new Error('fetch claims failed')
  return res.json()
}

export async function adminLogout() {
  const res = await fetch(`${API_BASE}/admin/logout`, {
    method: 'POST',
    headers: headers(true),
  })
  if (!res.ok) throw new Error('logout failed')
  setSession('')
  return res.json()
}

export async function fetchUserIDs() {
  const res = await fetch(`${API_BASE}/admin/ids`, {
    headers: headers(true),
  })
  if (!res.ok) throw new Error('fetch ids failed')
  return res.json()
}

export async function addUserIDs(ids) {
  const res = await fetch(`${API_BASE}/admin/ids`, {
    method: 'POST',
    headers: headers(true),
    body: JSON.stringify({ ids }),
  })
  if (!res.ok) throw new Error('add ids failed')
  return res.json()
}

export async function deleteUserID(id) {
  const res = await fetch(`${API_BASE}/admin/ids/${id}`, {
    method: 'DELETE',
    headers: headers(true),
  })
  if (!res.ok) throw new Error('delete id failed')
  return res.json()
}

