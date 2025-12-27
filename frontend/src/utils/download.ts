/**
 * Download utility functions
 */

/**
 * Downloads a string as a text file in the browser.
 * @param content - The text content to download
 * @param filename - The filename for the downloaded file
 */
export function downloadTextFile(content: string, filename: string): void {
  const blob = new Blob([content], { type: 'text/plain' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
