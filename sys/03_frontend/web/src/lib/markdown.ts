// Plain-text excerpt derived from a Markdown body, for article cards and
// <meta name="description"> (dev-plan-09-frontend-learn 9.3/9.5). Articles
// have no separate summary field, so we derive one rather than add a column.
export function excerpt(markdown: string, length = 100): string {
  const plain = markdown
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`[^`]*`/g, '')
    .replace(/!\[[^\]]*\]\([^)]*\)/g, '')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/[#*_>~-]/g, '')
    .replace(/\s+/g, ' ')
    .trim()
  return plain.length > length ? plain.slice(0, length) + '…' : plain
}
