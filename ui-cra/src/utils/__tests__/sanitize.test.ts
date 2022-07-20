import { cleanHref } from '../sanitize';

describe('cleanHref', () => {
  it('should filter out scheme: javascript', () => {
    expect(cleanHref('javascript:alert(1);')).toBeUndefined();
  });

  it('should filter out scheme: data', () => {
    expect(
      cleanHref('data:text/html,<script>alert(document.domain)</script>'),
    ).toBeUndefined();
  });

  it('should handle empty strings', () => {
    expect(cleanHref('')).toBeUndefined();
  });

  it('should handle undefined', () => {
    expect(cleanHref(undefined)).toBeUndefined();
  });

  it('should allow valid http/https links', () => {
    [
      'https://example.com/',
      'http://example.com/',
      'https://example.com/path',
      'https://example.com/path?qs=foo',
      'https://example.com/path?qs=foo#bar',
    ].forEach(v => {
      expect(cleanHref(v)).toEqual(v);
    });
  });
});
