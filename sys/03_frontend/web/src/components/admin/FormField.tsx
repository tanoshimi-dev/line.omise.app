'use client'

// Shared form primitives for the admin CRUD forms (dev-plan-11-frontend-admin
// — "共通化できる部分は共通コンポーネント化").

interface FieldProps {
  label: string
  name: string
  value: string
  onChange: (value: string) => void
  required?: boolean
  placeholder?: string
}

export function TextField({ label, name, value, onChange, required, placeholder }: FieldProps) {
  return (
    <label className="block">
      <span className="text-sm font-medium text-gray-700">
        {label}
        {required && <span className="text-red-500"> *</span>}
      </span>
      <input
        type="text"
        name={name}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        placeholder={placeholder}
        className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-line-green focus:outline-none focus:ring-1 focus:ring-line-green"
      />
    </label>
  )
}

export function NumberField({ label, name, value, onChange }: { label: string; name: string; value: number; onChange: (value: number) => void }) {
  return (
    <label className="block">
      <span className="text-sm font-medium text-gray-700">{label}</span>
      <input
        type="number"
        name={name}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-line-green focus:outline-none focus:ring-1 focus:ring-line-green"
      />
    </label>
  )
}

export function TextAreaField({ label, name, value, onChange, required, rows = 10 }: FieldProps & { rows?: number }) {
  return (
    <label className="block">
      <span className="text-sm font-medium text-gray-700">
        {label}
        {required && <span className="text-red-500"> *</span>}
      </span>
      <textarea
        name={name}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        rows={rows}
        className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 font-mono text-sm focus:border-line-green focus:outline-none focus:ring-1 focus:ring-line-green"
      />
    </label>
  )
}

export function StatusSelect({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  return (
    <label className="block">
      <span className="text-sm font-medium text-gray-700">公開状態</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-line-green focus:outline-none focus:ring-1 focus:ring-line-green"
      >
        <option value="draft">下書き</option>
        <option value="published">公開</option>
      </select>
    </label>
  )
}

export function SubmitButton({ saving, children }: { saving: boolean; children: React.ReactNode }) {
  return (
    <button
      type="submit"
      disabled={saving}
      className="rounded-full bg-line-green px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-line-green-dark disabled:opacity-50"
    >
      {saving ? '保存中…' : children}
    </button>
  )
}
