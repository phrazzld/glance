'use strict';

/**
 * Promptfoo assertion helpers for Glance eval suite.
 *
 * Each function takes (output: string) and returns a Promptfoo assertion result:
 *   { pass: boolean, score: number, reason: string }
 *
 * Usage in promptfooconfig.yaml:
 *   assert:
 *     - type: javascript
 *       value: "const a = require('./assertions/glance.assert.js'); return a.hasRequiredHeaders(output);"
 */

/**
 * Verifies the output contains all three required section headers.
 * Enforces the Output format constraint in the v2 prompt.
 */
function hasRequiredHeaders(output) {
  const required = ['## Purpose', '## Key Roles', '## Dependencies and Caveats'];
  const missing = required.filter((h) => !output.includes(h));
  if (missing.length === 0) {
    return { pass: true, score: 1, reason: 'All required headers present' };
  }
  return {
    pass: false,
    score: 0,
    reason: `Missing required headers: ${missing.join(', ')}`,
  };
}

/**
 * Rejects output that describes CLI flags, options, or defaults.
 * The v2 prompt forbids this; config_pkg (which defines --force and --prompt-file)
 * is the key constraint test.
 */
function noCliSpeculation(output) {
  const checks = [
    { re: /--[a-z][-a-z0-9]*/g, label: 'CLI flag (--flag)' },
    { re: /\[default:/gi, label: '[default: ...]' },
    { re: /`-[a-zA-Z]/g, label: 'short flag in backticks' },
    { re: /\bcommand[- ]line (flag|option|argument)/gi, label: 'CLI flag mention' },
  ];
  const found = [];
  for (const { re, label } of checks) {
    if (re.test(output)) found.push(label);
  }
  if (found.length === 0) {
    return { pass: true, score: 1, reason: 'No CLI flag descriptions found' };
  }
  return {
    pass: false,
    score: 0,
    reason: `CLI speculation detected: ${found.join('; ')}`,
  };
}

/**
 * Rejects output containing recommendation or advisory language.
 * Both v1 and v2 prompts forbid recommendations, but v2 is more explicit.
 */
function noRecommendations(output) {
  const patterns = [
    /\bi recommend\b/i,
    /\byou should\b/i,
    /\bconsider using\b/i,
    /\bwould benefit from\b/i,
    /\bit would be (better|good|ideal)\b/i,
  ];
  const found = patterns.filter((p) => p.test(output));
  if (found.length === 0) {
    return { pass: true, score: 1, reason: 'No recommendation language found' };
  }
  return { pass: false, score: 0, reason: 'Recommendation language detected' };
}

/**
 * Rejects output containing speculative language about behavior or intent.
 * The v2 prompt explicitly requires omitting unverifiable claims.
 */
function noSpeculation(output) {
  const patterns = [
    /\blikely\b/i,
    /\bprobably\b/i,
    /\bit seems\b/i,
    /\bseems to\b/i,
    /\bmight be used\b/i,
  ];
  const found = patterns.filter((p) => p.test(output));
  if (found.length === 0) {
    return { pass: true, score: 1, reason: 'No speculative language found' };
  }
  return { pass: false, score: 0, reason: 'Speculative language detected' };
}

module.exports = { hasRequiredHeaders, noCliSpeculation, noRecommendations, noSpeculation };
