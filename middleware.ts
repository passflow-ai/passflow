import createMiddleware from 'next-intl/middleware';
import { routing } from './src/i18n/routing';

export default createMiddleware(routing);

export const config = {
  // Match all pathnames except for
  // - api routes
  // - _next (Next.js internals)
  // - _vercel (Vercel internals)
  // - static files (images, fonts, etc.)
  matcher: [
    '/((?!api|_next|_vercel|.*\\..*).*)',
    // However, match all pathnames within `/api/`, except for
    // - `/api/newsletter` (if you want to exclude some API routes)
  ]
};
