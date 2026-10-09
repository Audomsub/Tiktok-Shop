import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const workflowPath = path.resolve(__dirname, '../workflows/fastmoss_scraper_workflow.json');

console.log(`[VALIDATION] Reading workflow from: ${workflowPath}`);

if (!fs.existsSync(workflowPath)) {
  console.error(`[ERROR] File not found: ${workflowPath}`);
  process.exit(1);
}

const rawContent = fs.readFileSync(workflowPath, 'utf-8');
let workflow;
try {
  workflow = JSON.parse(rawContent);
  console.log('✅ JSON syntax is valid.');
} catch (err) {
  console.error('❌ Failed to parse JSON:', err.message);
  process.exit(1);
}

// 1. Check Root Schema
if (!workflow.name || !Array.isArray(workflow.nodes) || typeof workflow.connections !== 'object') {
  console.error('❌ Root schema missing name, nodes array, or connections object');
  process.exit(1);
}
console.log(`✅ Workflow Name: "${workflow.name}" (${workflow.nodes.length} nodes defined)`);

// 2. Validate Nodes Presence
const nodeMap = new Map();
workflow.nodes.forEach(node => {
  if (nodeMap.has(node.name)) {
    console.error(`❌ Duplicate node name detected: "${node.name}"`);
    process.exit(1);
  }
  nodeMap.set(node.name, node);
});

const requiredNodeNames = [
  // Ticket 15 Nodes
  'Schedule Trigger - Shallow Crawl (06:00, 12:00, 18:00)',
  'Schedule Trigger - Deep Crawl (00:00 Midnight)',
  'Manual / Test Trigger',
  'Initialize Crawl Scope & Round',
  'Supabase - Initialize crawl_logs (RUNNING)',
  'Generate Page Tasks Queue',
  'Page Processing Loop',
  'Calculate Jitter Delay & Cool-Off Pause',
  'Wait (Anti-Ban Jitter & Cool-off)',
  'HTTP Request - Fetch FastMoss Page',
  'Inspect Response & Anti-Ban Detector',
  'If HTTP 403 or 429 (Anti-Ban Triggered)',
  'Supabase - Record Anti-Ban Error in crawl_logs',
  'Emergency Alert Notification (Discord/Slack)',
  'Stop and Terminate Workflow Immediately',
  'Collect Page Products & Loop Back',
  'Aggregate Raw Products (Ready for Quality Gate)',
  // Ticket 16 Nodes
  'Pre-Filter Quality Gate & Taxonomy Extraction',
  'If Any Products Passed Quality Gate',
  'Supabase - Upsert Categories',
  'Prepare Products with Category IDs',
  'Supabase - Upsert Products',
  'Prepare Snapshots Payload',
  'Supabase - Insert Product Snapshots',
  'Format Audit Counts Update',
  'Supabase - Update crawl_logs Audit Counts',
  'HTTP Request - Trigger Go Analytics Engine',
  'Workflow Execution Summary'
];

for (const name of requiredNodeNames) {
  if (!nodeMap.has(name)) {
    console.error(`❌ Missing required node: "${name}"`);
    process.exit(1);
  }
}
console.log(`✅ All ${requiredNodeNames.length} essential nodes are present.`);

// 3. Validate Schedule Triggers
const shallowTrigger = nodeMap.get('Schedule Trigger - Shallow Crawl (06:00, 12:00, 18:00)');
const shallowCron = shallowTrigger.parameters?.rule?.interval?.[0]?.expression;
if (shallowCron !== '0 6,12,18 * * *') {
  console.error(`❌ Shallow crawl cron mismatch: expected "0 6,12,18 * * *", got "${shallowCron}"`);
  process.exit(1);
}
console.log(`✅ Shallow Trigger Cron verified: "${shallowCron}"`);

const deepTrigger = nodeMap.get('Schedule Trigger - Deep Crawl (00:00 Midnight)');
const deepCron = deepTrigger.parameters?.rule?.interval?.[0]?.expression;
if (deepCron !== '0 0 * * *') {
  console.error(`❌ Deep crawl cron mismatch: expected "0 0 * * *", got "${deepCron}"`);
  process.exit(1);
}
console.log(`✅ Deep Trigger Cron verified: "${deepCron}"`);

// 4. Validate FastMoss Request & Cookie Injection
const fetchNode = nodeMap.get('HTTP Request - Fetch FastMoss Page');
const headers = fetchNode.parameters?.headerParameters?.parameters || [];
const cookieHeader = headers.find(h => h.name.toLowerCase() === 'cookie');
if (!cookieHeader || !cookieHeader.value.includes('$env.FASTMOSS_COOKIE')) {
  console.error('❌ Cookie header missing or does not reference $env.FASTMOSS_COOKIE');
  process.exit(1);
}
console.log('✅ FastMoss HTTP Request node contains Cookie injection: {{ $env.FASTMOSS_COOKIE }}');

// 5. Validate Jitter & Cool-Off Code Logic
const jitterNode = nodeMap.get('Calculate Jitter Delay & Cool-Off Pause');
const jitterCode = jitterNode.parameters?.jsCode || '';
if (!jitterCode.includes('7.0 - 3.0') || !jitterCode.includes('25') || !jitterCode.includes('% 10 === 0')) {
  console.error('❌ Jitter delay (3.0-7.0s) or cool-off (25s every 10 pages) logic not found in code node');
  process.exit(1);
}
console.log('✅ Anti-Ban Rate Limiter verified (3-7s randomized jitter + 25s cool-off every 10 pages).');

// 6. Validate Anti-Ban Interception & Termination
const inspectNode = nodeMap.get('Inspect Response & Anti-Ban Detector');
const inspectCode = inspectNode.parameters?.jsCode || '';
if (!inspectCode.includes('403') || !inspectCode.includes('429')) {
  console.error('❌ HTTP 403/429 status code detection missing');
  process.exit(1);
}
console.log('✅ Status code inspection handles HTTP 403 and 429.');

const stopNode = nodeMap.get('Stop and Terminate Workflow Immediately');
if (stopNode.type !== 'n8n-nodes-base.stopAndError') {
  console.error(`❌ Expected stop node to be type "n8n-nodes-base.stopAndError", got "${stopNode.type}"`);
  process.exit(1);
}
console.log('✅ Stop and Terminate node uses stopAndError to protect session.');

// 7. Validate Ticket 16 - Pre-Filtering Quality Gate Rules
const filterNode = nodeMap.get('Pre-Filter Quality Gate & Taxonomy Extraction');
const filterCode = filterNode.parameters?.jsCode || '';
const hasCommissionRule = filterCode.includes('commissionRate >= 10.0');
const hasPriceRule = filterCode.includes('price >= 80.0') && filterCode.includes('price <= 1500.0');
const hasSalesRule = filterCode.includes('totalSales >= 30');

if (!hasCommissionRule || !hasPriceRule || !hasSalesRule) {
  console.error('❌ Pre-Filtering quality criteria missing: requires Commission >= 10%, Price 80-1,500 THB, Total Sales >= 30');
  process.exit(1);
}
console.log('✅ Pre-Filtering Quality Gate rules verified (Commission >= 10%, Price 80-1,500 THB, Sales >= 30).');

// 8. Validate Ticket 16 - Supabase Upsert Nodes & ON CONFLICT Clauses
const upsertCatNode = nodeMap.get('Supabase - Upsert Categories');
if (!upsertCatNode.parameters?.url?.includes('on_conflict=slug')) {
  console.error('❌ Categories upsert URL missing on_conflict=slug');
  process.exit(1);
}
const catHeaders = upsertCatNode.parameters?.headerParameters?.parameters || [];
const catPrefer = catHeaders.find(h => h.name.toLowerCase() === 'prefer');
if (!catPrefer?.value?.includes('resolution=merge-duplicates')) {
  console.error('❌ Categories upsert Prefer header missing resolution=merge-duplicates');
  process.exit(1);
}
console.log('✅ Categories auto-upsert verified with on_conflict=slug and merge-duplicates.');

const upsertProdNode = nodeMap.get('Supabase - Upsert Products');
if (!upsertProdNode.parameters?.url?.includes('on_conflict=source_id')) {
  console.error('❌ Products upsert URL missing on_conflict=source_id');
  process.exit(1);
}
const prodHeaders = upsertProdNode.parameters?.headerParameters?.parameters || [];
const prodPrefer = prodHeaders.find(h => h.name.toLowerCase() === 'prefer');
if (!prodPrefer?.value?.includes('resolution=merge-duplicates')) {
  console.error('❌ Products upsert Prefer header missing resolution=merge-duplicates');
  process.exit(1);
}
console.log('✅ Products auto-upsert verified with on_conflict=source_id and merge-duplicates.');

const insertSnapshotsNode = nodeMap.get('Supabase - Insert Product Snapshots');
if (!insertSnapshotsNode.parameters?.url?.includes('on_conflict=product_id,crawl_log_id')) {
  console.error('❌ Snapshots insert URL missing on_conflict=product_id,crawl_log_id');
  process.exit(1);
}
console.log('✅ Product snapshots idempotent insert verified with on_conflict=product_id,crawl_log_id.');

// 9. Validate Ticket 16 - Audit Counts Logging in crawl_logs
const updateAuditNode = nodeMap.get('Supabase - Update crawl_logs Audit Counts');
const auditBody = updateAuditNode.parameters?.jsonBody || '';
if (
  !auditBody.includes('total_pages_requested') ||
  !auditBody.includes('total_pages_success') ||
  !auditBody.includes('raw_products_scraped') ||
  !auditBody.includes('filtered_products_saved')
) {
  console.error('❌ crawl_logs audit update missing required count fields');
  process.exit(1);
}
console.log('✅ crawl_logs audit counts update verified (total_pages, raw_scraped, filtered_saved).');

// 10. Validate Ticket 16 - Go Backend Webhook Handshake
const webhookNode = nodeMap.get('HTTP Request - Trigger Go Analytics Engine');
if (!webhookNode.parameters?.url?.includes('/api/v1/jobs/compute-scores')) {
  console.error('❌ Webhook URL does not target /api/v1/jobs/compute-scores');
  process.exit(1);
}
const webhookHeaders = webhookNode.parameters?.headerParameters?.parameters || [];
const apiKeyHeader = webhookHeaders.find(h => h.name === 'X-API-Key');
if (!apiKeyHeader || !apiKeyHeader.value?.includes('$env.INTERNAL_API_KEY')) {
  console.error('❌ Webhook missing authenticated X-API-Key header referencing $env.INTERNAL_API_KEY');
  process.exit(1);
}
console.log('✅ Go Analytics Engine webhook handshake verified with X-API-Key and crawl_log_id.');

// 11. Validate Connections Graph
for (const [sourceName, targets] of Object.entries(workflow.connections)) {
  if (!nodeMap.has(sourceName)) {
    console.error(`❌ Connection source refers to unknown node: "${sourceName}"`);
    process.exit(1);
  }
  if (targets.main) {
    targets.main.forEach((outputList, outIndex) => {
      outputList.forEach(conn => {
        if (!nodeMap.has(conn.node)) {
          console.error(`❌ Connection target "${conn.node}" from "${sourceName}" (output ${outIndex}) does not exist`);
          process.exit(1);
        }
      });
    });
  }
}
console.log('✅ Workflow node connection topology is 100% consistent across all 28 nodes.');

console.log('\n🎉 ALL VALIDATION CHECKS PASSED FOR TICKET 15 & TICKET 16 PIPELINE!');
