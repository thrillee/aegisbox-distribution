'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'
import { Code, AlertCircle, CheckCircle, Eye, EyeOff } from 'lucide-react'

interface JsonEditorProps {
  value: Record<string, any> | null
  onChange: (value: Record<string, any> | null) => void
  className?: string
  label?: string
  description?: string
}

export function JsonEditor({
  value,
  onChange,
  className,
  label = 'Provider Config (JSON)',
  description = 'Additional provider-specific configuration in JSON format',
}: JsonEditorProps) {
  const [inputValue, setInputValue] = React.useState(() =>
    value ? JSON.stringify(value, null, 2) : ''
  )
  const [error, setError] = React.useState<string | null>(null)
  const [showPreview, setShowPreview] = React.useState(false)

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newValue = e.target.value
    setInputValue(newValue)

    if (!newValue.trim()) {
      onChange(null)
      setError(null)
      return
    }

    try {
      const parsed = JSON.parse(newValue)
      onChange(parsed)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid JSON')
    }
  }

  const parsedValue = React.useMemo(() => {
    if (!inputValue.trim()) return null
    try {
      return JSON.parse(inputValue)
    } catch {
      return null
    }
  }, [inputValue])

  return (
    <div className={cn('space-y-3', className)}>
      <div className="flex items-center justify-between">
        <div>
          <label className="block text-sm font-medium text-zinc-400 mb-1">
            {label}
          </label>
          <p className="text-xs text-zinc-500">{description}</p>
        </div>
        {parsedValue && (
          <button
            type="button"
            onClick={() => setShowPreview(!showPreview)}
            className="flex items-center gap-1.5 px-2 py-1 text-xs text-zinc-400 hover:text-white bg-zinc-800 hover:bg-zinc-700 rounded-md transition-colors"
          >
            {showPreview ? (
              <>
                <EyeOff className="h-3.5 w-3.5" />
                Hide Preview
              </>
            ) : (
              <>
                <Eye className="h-3.5 w-3.5" />
                Show Preview
              </>
            )}
          </button>
        )}
      </div>

      <div className="relative">
        <textarea
          value={inputValue}
          onChange={handleInputChange}
          placeholder='{"key": "value"}'
          className={cn(
            'w-full h-48 px-3 py-2 bg-zinc-800 border rounded-lg text-white font-mono text-sm',
            'placeholder:text-zinc-500 resize-none focus:outline-none focus:ring-2 focus:ring-white/20',
            error
              ? 'border-red-500/50 focus:border-red-500'
              : 'border-white/10 focus:border-white/20'
          )}
        />

        {error && (
          <div className="absolute bottom-3 right-3 flex items-center gap-1.5 px-2 py-1 bg-red-500/10 border border-red-500/20 rounded text-xs text-red-400">
            <AlertCircle className="h-3.5 w-3.5" />
            {error}
          </div>
        )}

        {!error && parsedValue && (
          <div className="absolute bottom-3 right-3 flex items-center gap-1.5 px-2 py-1 bg-emerald-500/10 border border-emerald-500/20 rounded text-xs text-emerald-400">
            <CheckCircle className="h-3.5 w-3.5" />
            Valid JSON
          </div>
        )}

        <div className="absolute top-2 right-2 pointer-events-none">
          <Code className="h-4 w-4 text-zinc-500" />
        </div>
      </div>

      {showPreview && parsedValue && (
        <div className="bg-zinc-900 border border-white/10 rounded-lg p-4">
          <div className="flex items-center gap-2 mb-2 text-xs text-zinc-400 uppercase tracking-wider">
            <Code className="h-3.5 w-3.5" />
            Preview
          </div>
          <pre className="text-xs text-emerald-400 font-mono overflow-auto max-h-64">
            {JSON.stringify(parsedValue, null, 2)}
          </pre>
        </div>
      )}
    </div>
  )
}
