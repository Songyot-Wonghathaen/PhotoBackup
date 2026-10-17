// Dev-mode shim: Nuxt's SPA fallback answers unknown paths with index.html (200),
// which hides these routes from the Wails dev asset proxy. Returning 404 lets Wails
// fall through to the Go image handler (/api/thumbnail).
export default defineEventHandler(() => {
  throw createError({ statusCode: 404, statusMessage: 'Not Found' })
})
