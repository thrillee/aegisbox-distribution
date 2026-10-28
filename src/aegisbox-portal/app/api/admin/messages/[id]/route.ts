import { db } from '@/lib/db'
import { messages, messageSegments, serviceProviders, spCredentials, mnos } from '@/lib/schema'
import { NextResponse } from 'next/server'
import { eq } from 'drizzle-orm'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const resolvedParams = await params
    const messageId = parseInt(resolvedParams.id)

    if (isNaN(messageId)) {
      return NextResponse.json({ error: 'Invalid message ID' }, { status: 400 })
    }

    const [message] = await db
      .select({
        id: messages.id,
        clientMessageId: messages.clientMessageId,
        clientRef: messages.clientRef,
        originalSourceAddr: messages.originalSourceAddr,
        originalDestinationAddr: messages.originalDestinationAddr,
        shortMessage: messages.shortMessage,
        totalSegments: messages.totalSegments,
        submittedAt: messages.submittedAt,
        receivedAt: messages.receivedAt,
        completedAt: messages.completedAt,
        processingStatus: messages.processingStatus,
        finalStatus: messages.finalStatus,
        errorCode: messages.errorCode,
        errorDescription: messages.errorDescription,
        cost: messages.cost,
        currencyCode: messages.currencyCode,
        retryCount: messages.retryCount,
        serviceProviderId: messages.serviceProviderId,
        spName: serviceProviders.name,
        spCredentialId: messages.spCredentialId,
        spSystemId: spCredentials.systemId,
        routedMnoId: messages.routedMnoId,
        mnoName: mnos.name,
        mnoCountryCode: mnos.countryCode,
      })
      .from(messages)
      .leftJoin(serviceProviders, eq(messages.serviceProviderId, serviceProviders.id))
      .leftJoin(spCredentials, eq(messages.spCredentialId, spCredentials.id))
      .leftJoin(mnos, eq(messages.routedMnoId, mnos.id))
      .where(eq(messages.id, messageId))
      .limit(1)

    if (!message) {
      return NextResponse.json({ error: 'Message not found' }, { status: 404 })
    }

    const segments = await db
      .select({
        id: messageSegments.id,
        segmentSeqn: messageSegments.segmentSeqn,
        mnoMessageId: messageSegments.mnoMessageId,
        mnoConnectionId: messageSegments.mnoConnectionId,
        sentToMnoAt: messageSegments.sentToMnoAt,
        dlrStatus: messageSegments.dlrStatus,
        dlrReceivedAt: messageSegments.dlrReceivedAt,
        errorCode: messageSegments.errorCode,
        errorDescription: messageSegments.errorDescription,
        dlrSource: messageSegments.dlrSource,
        createdAt: messageSegments.createdAt,
      })
      .from(messageSegments)
      .where(eq(messageSegments.messageId, messageId))
      .orderBy(messageSegments.segmentSeqn)

    return NextResponse.json({
      message,
      segments,
    })
  } catch (error) {
    console.error('[v0] Error fetching message:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
