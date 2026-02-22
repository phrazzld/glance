'use strict';

const fs = require('fs');
const path = require('path');

/**
 * Custom Promptfoo provider for Glance evals.
 *
 * Receives vars.fixtureFile (path to fixture JSON relative to evals/),
 * renders the Go prompt template in JS, calls OpenRouter, and returns output.
 *
 * Provider-level config (from bakeoff.yaml providers[].config):
 *   promptVersion: 'v1' | 'v2'  (default: 'v2')
 *
 * Test-level vars override provider config when vars.promptVersion is set.
 */
class GlanceProvider {
  constructor(options = {}) {
    this.promptVersion = options.promptVersion || 'v2';
  }

  id() {
    return `glance-provider-${this.promptVersion}`;
  }

  async callApi(prompt, context) {
    const vars = context?.vars || {};
    const fixtureFile = vars.fixtureFile;
    if (!fixtureFile) {
      throw new Error('vars.fixtureFile is required');
    }

    // Load fixture
    const fixturePath = path.resolve(__dirname, fixtureFile);
    const fixture = JSON.parse(fs.readFileSync(fixturePath, 'utf8'));

    // Format file contents — mirrors Go's FormatFileContents:
    // "=== file: {name} ===\n{content}\n\n"
    let fileContents = '';
    for (const [filename, content] of Object.entries(fixture.files || {})) {
      fileContents += `=== file: ${filename} ===\n${content}\n\n`;
    }

    // Format sub-glances (array of strings from child .glance.md files)
    const subGlances = (fixture.subGlances || []).join('\n\n---\n\n');

    // Load prompt template (v1 or v2)
    const version = vars.promptVersion || this.promptVersion;
    const templatePath = path.resolve(__dirname, `prompts/${version}.txt`);
    const template = fs.readFileSync(templatePath, 'utf8');

    // Render template — replicate Go text/template variable substitution
    const renderedPrompt = template
      .replace(/\{\{\.Directory\}\}/g, fixture.directory || '')
      .replace(/\{\{\.SubGlances\}\}/g, subGlances)
      .replace(/\{\{\.FileContents\}\}/g, fileContents);

    // Call OpenRouter
    const apiKey = process.env.OPENROUTER_API_KEY;
    if (!apiKey) {
      throw new Error('OPENROUTER_API_KEY environment variable is not set');
    }

    const response = await fetch('https://openrouter.ai/api/v1/chat/completions', {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
        'HTTP-Referer': 'https://github.com/phrazzld/glance',
        'X-Title': 'Glance Evals',
      },
      body: JSON.stringify({
        model: 'google/gemini-2.0-flash-001',
        messages: [{ role: 'user', content: renderedPrompt }],
        temperature: 0,
        max_tokens: 1024,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`OpenRouter API error ${response.status}: ${errorText}`);
    }

    const data = await response.json();
    const output = data.choices?.[0]?.message?.content;

    if (!output) {
      throw new Error(`No output from API. Response: ${JSON.stringify(data)}`);
    }

    return { output };
  }
}

module.exports = GlanceProvider;
