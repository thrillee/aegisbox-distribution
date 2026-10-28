import { redirect } from 'next/navigation'

// The provider's wallet, balance and transactions are on the Wallet tab of the provider page.
export default async function ProviderWalletPage({ params }: Readonly<{ params: Promise<{ id: string }> }>) {
  const { id } = await params
  redirect(`/admin/service-providers/${id}`)
}
