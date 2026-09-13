import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

// Body format decided in dev-plan-05-content-api 5.4: Markdown. No
// interactivity needed, so this renders server-side. `prose` classes come
// from @tailwindcss/typography.
export default function Markdown({ children }: { children: string }) {
  return (
    <div className="prose prose-gray max-w-none prose-headings:font-bold prose-a:text-line-green">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{children}</ReactMarkdown>
    </div>
  )
}
