import { db } from '@/lib/db'
import { mnoConnections } from '@/lib/schema'
import { NextResponse } from 'next/server'

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const {
      mnoId,
      protocol,
      status,
      priority,
      systemId,
      password,
      host,
      port,
      useTls,
      bindType,
      systemType,
      enquireLinkIntervalSecs,
      requestTimeoutSecs,
      connectRetryDelaySecs,
      maxWindowSize,
      defaultDataCoding,
      sourceAddrTon,
      sourceAddrNpi,
      destAddrTon,
      destAddrNpi,
      httpConfig,
      providerConfig,
    } = body

    if (!mnoId || !protocol) {
      return NextResponse.json({ error: 'MNO ID and protocol are required' }, { status: 400 })
    }

    const result = await db.insert(mnoConnections).values({
      mnoId,
      protocol,
      status: status || 'active',
      priority: priority || 1,
      systemId: systemId || null,
      password: password || null,
      host: host || null,
      port: port || null,
      useTls: useTls || false,
      bindType: bindType || 'trx',
      systemType: systemType || null,
      enquireLinkIntervalSecs: enquireLinkIntervalSecs || 30,
      requestTimeoutSecs: requestTimeoutSecs || 10,
      connectRetryDelaySecs: connectRetryDelaySecs || 5,
      maxWindowSize: maxWindowSize || 10,
      defaultDataCoding: defaultDataCoding || 0,
      sourceAddrTon: sourceAddrTon || 1,
      sourceAddrNpi: sourceAddrNpi || 1,
      destAddrTon: destAddrTon || 1,
      destAddrNpi: destAddrNpi || 1,
      baseUrl: httpConfig?.endpoint_url || '',
      timeoutSecs: httpConfig?.timeout_secs || 30,
      providerConfig: providerConfig || httpConfig || null,
    }).returning()

    return NextResponse.json({ success: true, connection: result[0] })
  } catch (error) {
    console.error('[v0] Error creating MNO connection:', error)
    return NextResponse.json({ error: 'Failed to create connection' }, { status: 500 })
  }
}
