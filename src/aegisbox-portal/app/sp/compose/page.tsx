import { redirect } from 'next/navigation'

// Messages are sent over SMPP or the HTTP API; this page had placeholder content only.
export default function Page() {
  redirect('/sp/api-keys')
}
