import * as DOMPurify from 'dompurify';
//
// This should clean "javascript:"" and "data:" schemes which can both be
// used to craft malicious links:
// - "javascript:alert(1);"
// - "data:text/html,<script>alert(document.domain)</script>"
//
// It leaves relative links in tact (like "/flux_runtime")
//
// DOMPurify has a slightly awkward API but is supposedly one of the more
// battle-tested sanitization libraries.
//
export const cleanHref = (href: string | undefined): string | undefined => {
  if (!href) {
    return undefined;
  }
  const a = document.createElement('a');
  a.href = href;
  const cleanAnchor = DOMPurify.sanitize(a, { RETURN_DOM: true });
  // return undefined as "href: string | undefined"
  return cleanAnchor?.querySelector('a')?.href || undefined;
};
