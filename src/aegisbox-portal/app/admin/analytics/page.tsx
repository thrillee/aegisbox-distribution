import { redirect } from 'next/navigation'

// This page showed random sample data; the real platform analytics live under Messages.
export default function Page() {
  redirect('/admin/messages/analytics')
}
