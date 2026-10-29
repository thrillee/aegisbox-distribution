// Where customers connect, set per deployment.
export function smppConnectionInfo() {
  return {
    host: process.env.SMPP_PUBLIC_HOST || null,
    port: process.env.SMPP_PUBLIC_PORT || null,
  }
}

// Base URL of the gateway's HTTP API (customers POST to <url>/sms/incoming).
export function httpApiPublicUrl(): string | null {
  return process.env.HTTP_API_PUBLIC_URL?.replace(/\/+$/, '') || null
}
