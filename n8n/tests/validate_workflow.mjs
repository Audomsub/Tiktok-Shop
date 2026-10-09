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
  'Aggregate Raw Products (Ready for Quality Gate)'
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

// 7. Validate Connections Graph
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
console.log('✅ Workflow node connection topology is 100% consistent.');

console.log('\n🎉 ALL VALIDATION CHECKS PASSED FOR TICKET 15 WORKFLOW!');
