// Dev-mode shim: Nuxt's SPA fallback answers unknown paths with index.html (200),
// which hides these routes from the Wails dev asset proxy. Returning 404 lets Wails
// fall through to the Go image handler (/image/{id}, /thumb/{id}).
// In production builds (ssr:false + nuxt generate) this API route is not emitted,
// so the embedded asset server handles the paths directly.
export default defineEventHandler(() => {
  throw createError({ statusCode: 404, statusMessage: 'Not Found' })
})
