export function utf8_to_b64(str: string) {
  return window.btoa(unescape(encodeURIComponent(str)));
}

export function b64_to_utf8(str: string) {
  return decodeURIComponent(escape(window.atob(str)));
}

export const maybeFromBase64 = (data: string) => {
  try {
    return b64_to_utf8(data);
  } catch (err) {
    // FIXME: show a message to the user somehow.
    return '';
  }
};
