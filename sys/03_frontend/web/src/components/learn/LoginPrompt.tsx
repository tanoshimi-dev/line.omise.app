import { loginUrl } from '@/lib/api'

// Shown wherever a page requires login but the visitor isn't logged in
// (dev-plan-09-frontend-learn 9.2: viewing stays open to everyone, only
// account-specific actions require login).
export default function LoginPrompt({ message }: { message: string }) {
  return (
    <div className="rounded-2xl border border-gray-100 bg-gray-50 p-6 text-center">
      <p className="text-sm text-gray-600">{message}</p>
      <div className="mt-4 flex items-center justify-center gap-3">
        <a
          href={loginUrl('line')}
          className="rounded-full bg-line-green px-4 py-1.5 text-sm font-medium text-white hover:bg-line-green-dark transition-colors"
        >
          LINEでログイン
        </a>
        <a
          href={loginUrl('google')}
          className="rounded-full border border-gray-300 px-4 py-1.5 text-sm font-medium text-gray-700 hover:border-gray-400 transition-colors"
        >
          Googleでログイン
        </a>
      </div>
    </div>
  )
}
