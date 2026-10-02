// shared/lib/a11y/announce — write to a hidden aria-live region so screen
// readers pick up dynamic state changes (e.g. "Filter applied", "3 specs
// loaded"). Mounts a singleton <div role="status"> on first call.

let liveNode: HTMLDivElement | null = null;

function ensureNode(priority: 'polite' | 'assertive'): HTMLDivElement {
  if (liveNode && liveNode.getAttribute('aria-live') === priority) {
    return liveNode;
  }
  // Remove any previous live node (priority switch).
  if (liveNode) liveNode.remove();
  const node = document.createElement('div');
  node.setAttribute('role', 'status');
  node.setAttribute('aria-live', priority);
  node.setAttribute('aria-atomic', 'true');
  // Visually hidden but readable.
  node.style.cssText =
    'position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0;';
  document.body.appendChild(node);
  liveNode = node;
  return node;
}

/**
 * Announce a message to assistive tech. Safe to call before mount — the
 * aria-live div is created lazily on first call.
 *
 * @example
 *   announce('Loading 21 specs…');   // polite (default)
 *   announce('Server error 500',     // assertive (interrupts)
 *           'assertive');
 */
export function announce(message: string, priority: 'polite' | 'assertive' = 'polite'): void {
  if (typeof document === 'undefined') return;
  const node = ensureNode(priority);
  // Clear + set on the next frame so repeated identical strings still fire.
  node.textContent = '';
  requestAnimationFrame(() => {
    if (liveNode) liveNode.textContent = message;
  });
}