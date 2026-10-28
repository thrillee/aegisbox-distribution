import { db } from '@/lib/db'
import { mnoConnections } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const connId = parseInt(resolvedParams.id)

    if (isNaN(connId)) {
      return NextResponse.json({ error: 'Invalid connection ID' }, { status: 400 })
    }

    const [connection] = await db
      .select({
        id: mnoConnections.id,
        mnoId: mnoConnections.mnoId,
        protocol: mnoConnections.protocol,
        status: mnoConnections.status,
        priority: mnoConnections.priority,
        systemId: mnoConnections.systemId,
        host: mnoConnections.host,
        port: mnoConnections.port,
        useTls: mnoConnections.useTls,
        bindType: mnoConnections.bindType,
        systemType: mnoConnections.systemType,
        enquireLinkIntervalSecs: mnoConnections.enquireLinkIntervalSecs,
        requestTimeoutSecs: mnoConnections.requestTimeoutSecs,
        connectRetryDelaySecs: mnoConnections.connectRetryDelaySecs,
        maxWindowSize: mnoConnections.maxWindowSize,
        defaultDataCoding: mnoConnections.defaultDataCoding,
        sourceAddrTon: mnoConnections.sourceAddrTon,
        sourceAddrNpi: mnoConnections.sourceAddrNpi,
        destAddrTon: mnoConnections.destAddrTon,
        destAddrNpi: mnoConnections.destAddrNpi,
        baseUrl: mnoConnections.baseUrl,
        timeoutSecs: mnoConnections.timeoutSecs,
        providerConfig: mnoConnections.providerConfig,
        createdAt: mnoConnections.createdAt,
        updatedAt: mnoConnections.updatedAt,
      })
      .from(mnoConnections)
      .where(eq(mnoConnections.id, connId))
      .limit(1)

    if (!connection) {
      return NextResponse.json({ error: 'Connection not found' }, { status: 404 })
    }

    return NextResponse.json({ connection })
  } catch (error) {
    console.error('[v0] Error fetching connection:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const connId = parseInt(resolvedParams.id)

    if (isNaN(connId)) {
      return NextResponse.json({ error: 'Invalid connection ID' }, { status: 400 })
    }

    const body = await request.json()
    const {
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
      baseUrl,
      timeoutSecs,
      httpConfig,
      providerConfig,
    } = body

    const updateData: Record<string, any> = {
      updatedAt: new Date(),
    }

    if (status !== undefined) updateData.status = status
    if (priority !== undefined) updateData.priority = priority
    if (systemId !== undefined) updateData.systemId = systemId
    if (password !== undefined) updateData.password = password
    if (host !== undefined) updateData.host = host
    if (port !== undefined) updateData.port = port
    if (useTls !== undefined) updateData.useTls = useTls
    if (bindType !== undefined) updateData.bindType = bindType
    if (systemType !== undefined) updateData.systemType = systemType
    if (enquireLinkIntervalSecs !== undefined) updateData.enquireLinkIntervalSecs = enquireLinkIntervalSecs
    if (requestTimeoutSecs !== undefined) updateData.requestTimeoutSecs = requestTimeoutSecs
    if (connectRetryDelaySecs !== undefined) updateData.connectRetryDelaySecs = connectRetryDelaySecs
    if (maxWindowSize !== undefined) updateData.maxWindowSize = maxWindowSize
    if (defaultDataCoding !== undefined) updateData.defaultDataCoding = defaultDataCoding
    if (sourceAddrTon !== undefined) updateData.sourceAddrTon = sourceAddrTon
    if (sourceAddrNpi !== undefined) updateData.sourceAddrNpi = sourceAddrNpi
    if (destAddrTon !== undefined) updateData.destAddrTon = destAddrTon
    if (destAddrNpi !== undefined) updateData.destAddrNpi = destAddrNpi
    if (baseUrl !== undefined) updateData.baseUrl = baseUrl
    if (timeoutSecs !== undefined) updateData.timeoutSecs = timeoutSecs
    if (httpConfig !== undefined) {
      updateData.providerConfig = httpConfig
      if (httpConfig.endpoint_url) updateData.baseUrl = httpConfig.endpoint_url
      if (httpConfig.timeout_secs) updateData.timeoutSecs = httpConfig.timeout_secs
    }

    if (providerConfig !== undefined) {
      updateData.providerConfig = providerConfig
    }

    const [connection] = await db
      .update(mnoConnections)
      .set(updateData)
      .where(eq(mnoConnections.id, connId))
      .returning()

    if (!connection) {
      return NextResponse.json({ error: 'Connection not found' }, { status: 404 })
    }

    return NextResponse.json({ success: true, connection })
  } catch (error) {
    console.error('[v0] Error updating connection:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
