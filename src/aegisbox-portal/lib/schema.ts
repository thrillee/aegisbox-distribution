import { pgTable, serial, varchar, text, timestamp, boolean, integer, decimal, jsonb, pgEnum } from 'drizzle-orm/pg-core'
import { relations } from 'drizzle-orm'

export const statusEnum = pgEnum('status', ['pending', 'active', 'suspended', 'cancelled', 'inactive'])
export const protocolEnum = pgEnum('protocol', ['smpp', 'http'])
export const messageTypeEnum = pgEnum('message_type', ['sms', 'otp', 'flash', 'unicode', 'binary'])
export const transactionTypeEnum = pgEnum('transaction_type', ['debit', 'credit', 'reversal'])
export const bindTypeEnum = pgEnum('bind_type', ['trx', 'tx', 'rx'])
export const senderIdTypeEnum = pgEnum('sender_id_type', ['regular', 'otp'])

export const serviceProviders = pgTable('service_providers', {
  id: serial('id').primaryKey(),
  name: varchar('name', { length: 255 }).notNull(),
  email: varchar('email', { length: 255 }).notNull().unique(),
  status: statusEnum('status').default('pending').notNull(),
  defaultCurrencyCode: varchar('default_currency_code', { length: 3 }).notNull().default('NGN'),
  autoDlrEnabled: boolean('auto_dlr_enabled').notNull().default(false),
  autoDlrStatus: varchar('auto_dlr_status', { length: 20 }).notNull().default('DELIVRD'),
  autoDlrDelaySeconds: integer('auto_dlr_delay_seconds').notNull().default(300),
  autoDlrEnabledAt: timestamp('auto_dlr_enabled_at'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const spCredentials = pgTable('sp_credentials', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id').notNull(),
  protocol: protocolEnum('protocol').default('smpp').notNull(),
  status: statusEnum('status').default('active').notNull(),
  systemId: varchar('system_id', { length: 16 }),
  passwordHash: varchar('password_hash', { length: 255 }),
  bindType: bindTypeEnum('bind_type').default('trx'),
  apiKeyHash: varchar('api_key_hash', { length: 255 }),
  apiKeyIdentifier: varchar('api_key_identifier', { length: 64 }).unique(),
  httpConfig: jsonb('http_config'),
  scope: varchar('scope', { length: 50 }).notNull().default('default'),
  routingGroupId: integer('routing_group_id'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const mnos = pgTable('mnos', {
  id: serial('id').primaryKey(),
  name: varchar('name', { length: 100 }).notNull().unique(),
  countryCode: varchar('country_code', { length: 5 }).notNull(),
  networkCode: varchar('network_code', { length: 10 }),
  status: statusEnum('status').default('active').notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const mnoConnections = pgTable('mno_connections', {
  id: serial('id').primaryKey(),
  mnoId: integer('mno_id').notNull(),
  protocol: protocolEnum('protocol').default('smpp').notNull(),
  status: statusEnum('status').default('active').notNull(),
  priority: integer('priority').default(1).notNull(),
  systemId: varchar('system_id', { length: 16 }),
  password: varchar('password', { length: 255 }),
  host: varchar('host', { length: 255 }),
  port: integer('port'),
  useTls: boolean('use_tls').default(false),
  bindType: bindTypeEnum('bind_type').default('trx'),
  systemType: varchar('system_type', { length: 13 }),
  enquireLinkIntervalSecs: integer('enquire_link_interval_secs').default(30),
  requestTimeoutSecs: integer('request_timeout_secs').default(10),
  connectRetryDelaySecs: integer('connect_retry_delay_secs').default(5),
  maxWindowSize: integer('max_window_size').default(10),
  defaultDataCoding: integer('default_data_coding').default(0),
  sourceAddrTon: integer('source_addr_ton').default(1),
  sourceAddrNpi: integer('source_addr_npi').default(1),
  destAddrTon: integer('dest_addr_ton').default(1),
  destAddrNpi: integer('dest_addr_npi').default(1),
  apiKey: text('api_key'),
  baseUrl: varchar('base_url', { length: 255 }).default('').notNull(),
  username: varchar('username', { length: 255 }),
  httpPassword: text('http_password'),
  secretKey: text('secret_key'),
  defaultSender: varchar('default_sender', { length: 100 }),
  rateLimit: integer('rate_limit').default(100),
  email: varchar('email', { length: 255 }),
  supportsWebhook: boolean('supports_webhook').default(true),
  webhookPath: varchar('webhook_path', { length: 255 }),
  timeoutSecs: integer('timeout_secs').default(30),
  providerConfig: jsonb('provider_config'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const senderIds = pgTable('sender_ids', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id'),
  senderIdString: varchar('sender_id_string', { length: 16 }).notNull(),
  mnoConnectionId: integer('mno_connection_id'),
  status: statusEnum('status').default('pending').notNull(),
  providerId: integer('provider_id'),
  rateLimit: integer('rate_limit').default(100),
  otpTemplate: text('otp_template'),
  networks: text('networks').array().default({} as any),
  maxUsageCount: integer('max_usage_count').default(10000),
  currentUsageCount: integer('current_usage_count').default(0),
  resetIntervalHours: integer('reset_interval_hours'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const messages = pgTable('messages', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id').notNull(),
  spCredentialId: integer('sp_credential_id').notNull(),
  clientMessageId: varchar('client_message_id', { length: 64 }),
  clientRef: varchar('client_ref', { length: 64 }),
  originalSourceAddr: varchar('original_source_addr', { length: 21 }).notNull(),
  // Sender actually submitted upstream when a route sender pool replaced the original.
  sentSourceAddr: varchar('sent_source_addr', { length: 21 }),
  originalDestinationAddr: varchar('original_destination_addr', { length: 21 }).notNull(),
  shortMessage: text('short_message').notNull(),
  totalSegments: integer('total_segments').default(1).notNull(),
  submittedAt: timestamp('submitted_at'),
  receivedAt: timestamp('received_at').defaultNow().notNull(),
  approvedSenderId: integer('approved_sender_id'),
  templateId: integer('template_id'),
  routedMnoId: integer('routed_mno_id'),
  currencyCode: varchar('currency_code', { length: 3 }).notNull(),
  cost: decimal('cost', { precision: 19, scale: 6 }).default('0.000000'),
  processingStatus: varchar('processing_status', { length: 50 }).default('received').notNull(),
  finalStatus: varchar('final_status', { length: 50 }).default('pending').notNull(),
  errorCode: varchar('error_code', { length: 10 }),
  errorDescription: text('error_description'),
  processedForQueueAt: timestamp('processed_for_queue_at'),
  completedAt: timestamp('completed_at'),
  retryCount: integer('retry_count').default(0).notNull(),
  nextRetryAt: timestamp('next_retry_at'),
})

export const messageSegments = pgTable('message_segments', {
  id: serial('id').primaryKey(),
  messageId: integer('message_id').notNull(),
  segmentSeqn: integer('segment_seqn').notNull(),
  mnoMessageId: varchar('mno_message_id', { length: 64 }).unique(),
  mnoConnectionId: integer('mno_connection_id'),
  sentToMnoAt: timestamp('sent_to_mno_at'),
  dlrStatus: varchar('dlr_status', { length: 50 }),
  dlrReceivedAt: timestamp('dlr_received_at'),
  errorCode: varchar('error_code', { length: 10 }),
  errorDescription: text('error_description'),
  dlrSource: varchar('dlr_source', { length: 20 }),
  createdAt: timestamp('created_at').defaultNow().notNull(),
})

export const wallets = pgTable('wallets', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id').notNull(),
  currencyCode: varchar('currency_code', { length: 3 }).notNull(),
  balance: decimal('balance', { precision: 19, scale: 6 }).notNull().default('0.000000'),
  lowBalanceThreshold: decimal('low_balance_threshold', { precision: 19, scale: 6 }).notNull().default('0.000000'),
  lowBalanceNotifiedAt: timestamp('low_balance_notified_at'),
  walletType: varchar('wallet_type', { length: 20 }).notNull().default('prepaid'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const walletTransactions = pgTable('wallet_transactions', {
  id: serial('id').primaryKey(),
  walletId: integer('wallet_id').notNull(),
  messageId: integer('message_id'),
  transactionType: transactionTypeEnum('transaction_type').notNull(),
  amount: decimal('amount', { precision: 19, scale: 6 }).notNull(),
  balanceBefore: decimal('balance_before', { precision: 19, scale: 6 }).notNull(),
  balanceAfter: decimal('balance_after', { precision: 19, scale: 6 }).notNull(),
  description: text('description'),
  transactionDate: timestamp('transaction_date').defaultNow().notNull(),
  referenceTransactionId: integer('reference_transaction_id'),
})

export const pricingRules = pgTable('pricing_rules', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id').notNull(),
  mnoId: integer('mno_id'),
  currencyCode: varchar('currency_code', { length: 3 }).notNull(),
  pricePerSms: decimal('price_per_sms', { precision: 10, scale: 6 }).notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const routingGroups = pgTable('routing_groups', {
  id: serial('id').primaryKey(),
  name: varchar('name', { length: 255 }).notNull().unique(),
  description: text('description'),
  reference: varchar('reference', { length: 100 }).notNull().unique(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const routingAssignments = pgTable('routing_assignments', {
  id: serial('id').primaryKey(),
  routingGroupId: integer('routing_group_id').notNull(),
  msisdnPrefixGroupId: integer('msisdn_prefix_group_id').notNull(),
  mnoId: integer('mno_id').notNull(),
  status: statusEnum('status').default('active').notNull(),
  priority: integer('priority').default(0).notNull(),
  comment: text('comment'),
  weight: integer('weight').default(100).notNull(),
  maxRetryAttempts: integer('max_retry_attempts').default(2).notNull(),
  isAvailable: boolean('is_available').default(true).notNull(),
  consecutiveFailures: integer('consecutive_failures').default(0).notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const msisdnPrefixGroups = pgTable('msisdn_prefix_groups', {
  id: serial('id').primaryKey(),
  name: varchar('name', { length: 255 }).notNull().unique(),
  description: text('description'),
  reference: varchar('reference', { length: 100 }).notNull().unique(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const msisdnPrefixEntries = pgTable('msisdn_prefix_entries', {
  id: serial('id').primaryKey(),
  msisdnPrefixGroupId: integer('msisdn_prefix_group_id').notNull(),
  msisdnPrefix: varchar('msisdn_prefix', { length: 20 }).notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const templates = pgTable('templates', {
  id: serial('id').primaryKey(),
  serviceProviderId: integer('service_provider_id'),
  name: varchar('name', { length: 100 }).notNull(),
  content: text('content').notNull(),
  status: statusEnum('status').default('active').notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const otpAlternativeSenders = pgTable('otp_alternative_senders', {
  id: serial('id').primaryKey(),
  senderIdString: varchar('sender_id_string', { length: 16 }).notNull().unique(),
  serviceProviderId: integer('service_provider_id'),
  mnoId: integer('mno_id'),
  status: statusEnum('status').default('active').notNull(),
  currentUsageCount: integer('current_usage_count').notNull().default(0),
  maxUsageCount: integer('max_usage_count').notNull().default(10000),
  resetIntervalHours: integer('reset_interval_hours'),
  lastResetAt: timestamp('last_reset_at').defaultNow(),
  lastUsedAt: timestamp('last_used_at'),
  notes: text('notes'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const otpMessageTemplates = pgTable('otp_message_templates', {
  id: serial('id').primaryKey(),
  name: varchar('name', { length: 255 }).notNull().unique(),
  contentTemplate: text('content_template').notNull(),
  defaultBrandName: varchar('default_brand_name', { length: 100 }),
  status: statusEnum('status').default('active').notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const otpSenderTemplateAssignments = pgTable('otp_sender_template_assignments', {
  id: serial('id').primaryKey(),
  otpAlternativeSenderId: integer('otp_alternative_sender_id').notNull(),
  otpMessageTemplateId: integer('otp_message_template_id').notNull(),
  mnoId: integer('mno_id'),
  priority: integer('priority').default(0).notNull(),
  status: statusEnum('status').default('active').notNull(),
  assignmentUsageCount: decimal('assignment_usage_count', { precision: 19 }).notNull().default('0'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const deliveryReportsRaw = pgTable('delivery_reports_raw', {
  id: serial('id').primaryKey(),
  messageSegmentId: integer('message_segment_id'),
  mnoConnectionId: integer('mno_connection_id').notNull(),
  rawPayload: text('raw_payload'),
  receivedAt: timestamp('received_at').defaultNow().notNull(),
  processedAt: timestamp('processed_at'),
  processingStatus: statusEnum('processing_status').default('pending').notNull(),
})

export const dlrForwardingQueue = pgTable('dlr_forwarding_queue', {
  id: serial('id').primaryKey(),
  messageId: integer('message_id').notNull(),
  status: varchar('status', { length: 20 }).default('pending').notNull(),
  payload: jsonb('payload').notNull(),
  attempts: integer('attempts').default(0).notNull(),
  maxAttempts: integer('max_attempts').default(5).notNull(),
  lastAttemptAt: timestamp('last_attempt_at'),
  nextRetryAt: timestamp('next_retry_at'),
  errorMessage: text('error_message'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  lockedAt: timestamp('locked_at'),
  lockedBy: varchar('locked_by', { length: 100 }),
})

export const deadLetterQueue = pgTable('dead_letter_queue', {
  id: serial('id').primaryKey(),
  messageId: serial('message_id').notNull(),
  originalStatus: varchar('original_status', { length: 50 }),
  errorCode: varchar('error_code', { length: 50 }),
  errorDescription: text('error_description'),
  retryCount: integer('retry_count').default(0),
  failedAt: timestamp('failed_at').defaultNow(),
  createdAt: timestamp('created_at').defaultNow(),
  metadata: jsonb('metadata'),
})

export const orgMappings = pgTable('org_mappings', {
  id: serial('id').primaryKey(),
  clerkOrgId: varchar('clerk_org_id', { length: 64 }).notNull().unique(),
  serviceProviderId: integer('service_provider_id').notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const serviceProviderRelations = relations(serviceProviders, ({ many }) => ({
  credentials: many(spCredentials),
  wallets: many(wallets),
  messages: many(messages),
  senderIds: many(senderIds),
  pricingRules: many(pricingRules),
  templates: many(templates),
  otpAlternativeSenders: many(otpAlternativeSenders),
}))

export const mnosRelations = relations(mnos, ({ many }) => ({
  connections: many(mnoConnections),
}))

export const mnoConnectionsRelations = relations(mnoConnections, ({ one }) => ({
  mno: one(mnos, {
    fields: [mnoConnections.mnoId],
    references: [mnos.id],
  }),
}))

export const walletsRelations = relations(wallets, ({ one, many }) => ({
  serviceProvider: one(serviceProviders, {
    fields: [wallets.serviceProviderId],
    references: [serviceProviders.id],
  }),
  transactions: many(walletTransactions),
}))

export const walletTransactionsRelations = relations(walletTransactions, ({ one }) => ({
  wallet: one(wallets, {
    fields: [walletTransactions.walletId],
    references: [wallets.id],
  }),
}))

export const messagesRelations = relations(messages, ({ one, many }) => ({
  serviceProvider: one(serviceProviders, {
    fields: [messages.serviceProviderId],
    references: [serviceProviders.id],
  }),
  spCredential: one(spCredentials, {
    fields: [messages.spCredentialId],
    references: [spCredentials.id],
  }),
  senderId: one(senderIds, {
    fields: [messages.approvedSenderId],
    references: [senderIds.id],
  }),
  segments: many(messageSegments),
}))

export const messageSegmentsRelations = relations(messageSegments, ({ one }) => ({
  message: one(messages, {
    fields: [messageSegments.messageId],
    references: [messages.id],
  }),
}))

export const routingGroupsRelations = relations(routingGroups, ({ many }) => ({
  assignments: many(routingAssignments),
}))

export const msisdnPrefixGroupsRelations = relations(msisdnPrefixGroups, ({ many }) => ({
  entries: many(msisdnPrefixEntries),
}))

export const senderIdsRelations = relations(senderIds, ({ one }) => ({
  serviceProvider: one(serviceProviders, {
    fields: [senderIds.serviceProviderId],
    references: [serviceProviders.id],
  }),
}))

export const pricingRulesRelations = relations(pricingRules, ({ one }) => ({
  serviceProvider: one(serviceProviders, {
    fields: [pricingRules.serviceProviderId],
    references: [serviceProviders.id],
  }),
  mno: one(mnos, {
    fields: [pricingRules.mnoId],
    references: [mnos.id],
  }),
}))

export const templatesRelations = relations(templates, ({ one }) => ({
  serviceProvider: one(serviceProviders, {
    fields: [templates.serviceProviderId],
    references: [serviceProviders.id],
  }),
}))

export const otpMessageTemplatesRelations = relations(otpMessageTemplates, ({ many }) => ({
  assignments: many(otpSenderTemplateAssignments),
}))

export const otpAlternativeSendersRelations = relations(otpAlternativeSenders, ({ one, many }) => ({
  serviceProvider: one(serviceProviders, {
    fields: [otpAlternativeSenders.serviceProviderId],
    references: [serviceProviders.id],
  }),
  mno: one(mnos, {
    fields: [otpAlternativeSenders.mnoId],
    references: [mnos.id],
  }),
  assignments: many(otpSenderTemplateAssignments),
}))

export const otpSenderTemplateAssignmentsRelations = relations(otpSenderTemplateAssignments, ({ one }) => ({
  otpAlternativeSender: one(otpAlternativeSenders, {
    fields: [otpSenderTemplateAssignments.otpAlternativeSenderId],
    references: [otpAlternativeSenders.id],
  }),
  otpMessageTemplate: one(otpMessageTemplates, {
    fields: [otpSenderTemplateAssignments.otpMessageTemplateId],
    references: [otpMessageTemplates.id],
  }),
  mno: one(mnos, {
    fields: [otpSenderTemplateAssignments.mnoId],
    references: [mnos.id],
  }),
}))

export const orgMappingsRelations = relations(orgMappings, ({ one }) => ({
  serviceProvider: one(serviceProviders, {
    fields: [orgMappings.serviceProviderId],
    references: [serviceProviders.id],
  }),
}))

// Intelligent Routing Health Tracking Tables

export const routeHealth = pgTable('route_health', {
  id: serial('id').primaryKey(),
  routingAssignmentId: integer('routing_assignment_id').notNull().references(() => routingAssignments.id, { onDelete: 'cascade' }),
  senderId: integer('sender_id'),
  totalAttempts: integer('total_attempts').default(0).notNull(),
  successCount: integer('successful_attempts').default(0).notNull(),
  failCount: integer('failed_attempts').default(0).notNull(),
  successRate: decimal('success_rate', { precision: 5, scale: 2 }),
  healthScore: decimal('health_score', { precision: 5, scale: 2 }).default('100.00'),
  isAvailable: boolean('is_available').default(true).notNull(),
  unavailableSince: timestamp('unavailable_since'),
  unavailableReason: text('unavailable_reason'),
  autoUnavailable: boolean('auto_unavailable').default(false),
  consecutiveFailures: integer('consecutive_failures').default(0).notNull(),
  consecutiveSuccesses: integer('consecutive_successes').default(0).notNull(),
  lastAttemptAt: timestamp('last_attempt_at'),
  lastSuccessAt: timestamp('last_success_at'),
  lastFailureAt: timestamp('last_failure_at'),
  windowSize: integer('window_size').default(1000),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const senderIdHealth = pgTable('sender_id_health', {
  id: serial('id').primaryKey(),
  senderId: integer('sender_id').notNull().references(() => senderIds.id, { onDelete: 'cascade' }),
  mnoId: integer('mno_id').notNull().references(() => mnos.id, { onDelete: 'cascade' }),
  totalSent: integer('total_sent').default(0).notNull(),
  totalDelivered: integer('total_delivered').default(0).notNull(),
  totalFailed: integer('total_failed').default(0).notNull(),
  totalNoDlr: integer('total_no_dlr').default(0).notNull(),
  deliveryRate: decimal('delivery_rate', { precision: 5, scale: 2 }).default('100.00'),
  dlrResponseRate: decimal('dlr_response_rate', { precision: 5, scale: 2 }).default('100.00'),
  healthScore: decimal('health_score', { precision: 5, scale: 2 }).default('100.00'),
  isBlocked: boolean('is_blocked').default(false).notNull(),
  blockedAt: timestamp('blocked_at'),
  blockedReason: text('blocked_reason'),
  autoBlocked: boolean('auto_blocked').default(false),
  lastUnblockCheckAt: timestamp('last_unblock_check_at'),
  lastSuccessAt: timestamp('last_success_at'),
  lastFailureAt: timestamp('last_failure_at'),
  lastNoDlrAt: timestamp('last_no_dlr_at'),
  lastAttemptAt: timestamp('last_attempt_at'),
  windowSize: integer('window_size').default(1000),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const selectionStrategyEnum = pgEnum('selection_strategy', ['priority_weighted', 'health_based', 'round_robin'])

export const routingConfig = pgTable('routing_config', {
  id: serial('id').primaryKey(),
  routingGroupId: integer('routing_group_id').notNull().unique().references(() => routingGroups.id, { onDelete: 'cascade' }),
  selectionStrategy: selectionStrategyEnum('selection_strategy').default('priority_weighted').notNull(),
  maxRouteRetries: integer('max_route_retries').default(2).notNull(),
  dlrTimeoutSeconds: integer('dlr_timeout_seconds').default(300).notNull(),
  senderBlockThreshold: decimal('sender_block_threshold', { precision: 3, scale: 2 }).default('0.50').notNull(),
  senderUnblockThreshold: decimal('sender_unblock_threshold', { precision: 3, scale: 2 }).default('0.70').notNull(),
  routeAvailabilityThreshold: decimal('route_availability_threshold', { precision: 3, scale: 2 }).default('0.70').notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const healthRingBufferConfig = pgTable('health_ring_buffer_config', {
  id: serial('id').primaryKey(),
  healthCalculationWindow: integer('health_calculation_window').default(1000).notNull(),
  routeFailureThreshold: integer('route_failure_threshold').default(5).notNull(),
  routeRecoveryThreshold: integer('route_recovery_threshold').default(3).notNull(),
  senderAutoUnblockCheckIntervalMinutes: integer('sender_auto_unblock_check_interval_minutes').default(30).notNull(),
  dlrTimeoutCheckIntervalMinutes: integer('dlr_timeout_check_interval_minutes').default(5).notNull(),
  ringBufferCleanupIntervalHours: integer('ring_buffer_cleanup_interval_hours').default(24).notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const routeDeliveryAttempts = pgTable('route_delivery_attempts', {
  id: serial('id').primaryKey(),
  routingAssignmentId: integer('routing_assignment_id').notNull().references(() => routingAssignments.id, { onDelete: 'cascade' }),
  senderId: integer('sender_id'),
  messageId: integer('message_id'),
  success: boolean('success').notNull(),
  errorCode: varchar('error_code', { length: 10 }),
  errorDescription: text('error_description'),
  attemptedAt: timestamp('attempted_at').defaultNow().notNull(),
  recordedDate: timestamp('recorded_date').defaultNow().notNull(),
})

export const senderIdDeliveryAttempts = pgTable('sender_id_delivery_attempts', {
  id: serial('id').primaryKey(),
  senderId: varchar('sender_id', { length: 16 }).notNull(),
  mnoId: integer('mno_id').notNull().references(() => mnos.id, { onDelete: 'cascade' }),
  dlrReceived: boolean('dlr_received').notNull(),
  attemptedAt: timestamp('attempted_at').defaultNow().notNull(),
})

// Health Tracking Relations

export const routeHealthRelations = relations(routeHealth, ({ one }) => ({
  assignment: one(routingAssignments, {
    fields: [routeHealth.routingAssignmentId],
    references: [routingAssignments.id],
  }),
}))

export const senderIdHealthRelations = relations(senderIdHealth, ({ one }) => ({
  mno: one(mnos, {
    fields: [senderIdHealth.mnoId],
    references: [mnos.id],
  }),
}))

export const routingConfigRelations = relations(routingConfig, ({ one }) => ({
  routingGroup: one(routingGroups, {
    fields: [routingConfig.routingGroupId],
    references: [routingGroups.id],
  }),
}))

export const routeDeliveryAttemptsRelations = relations(routeDeliveryAttempts, ({ one }) => ({
  assignment: one(routingAssignments, {
    fields: [routeDeliveryAttempts.routingAssignmentId],
    references: [routingAssignments.id],
  }),
}))

export const senderIdDeliveryAttemptsRelations = relations(senderIdDeliveryAttempts, ({ one }) => ({
  mno: one(mnos, {
    fields: [senderIdDeliveryAttempts.mnoId],
    references: [mnos.id],
  }),
}))
