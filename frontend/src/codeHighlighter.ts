import {createBundledHighlighter, type HighlighterGeneric} from 'shiki/core';
import {createOnigurumaEngine} from 'shiki/engine/oniguruma';

const theme = 'github-dark';
const bundledLanguages = {
    bash: () => import('@shikijs/langs/bash'),
    css: () => import('@shikijs/langs/css'),
    go: () => import('@shikijs/langs/go'),
    html: () => import('@shikijs/langs/html'),
    javascript: () => import('@shikijs/langs/javascript'),
    json: () => import('@shikijs/langs/json'),
    markdown: () => import('@shikijs/langs/markdown'),
    powershell: () => import('@shikijs/langs/powershell'),
    typescript: () => import('@shikijs/langs/typescript'),
    yaml: () => import('@shikijs/langs/yaml')
} as const;

const bundledThemes = {
    [theme]: () => import('@shikijs/themes/github-dark')
} as const;

type HighlightLanguage = keyof typeof bundledLanguages;
type HighlightTheme = keyof typeof bundledThemes;
type HighlightHighlighter = HighlighterGeneric<HighlightLanguage, HighlightTheme>;

const createHighlighter = createBundledHighlighter({
    langs: bundledLanguages,
    themes: bundledThemes,
    engine: () => createOnigurumaEngine(() => import('shiki/wasm'))
});

const languageAliases: Record<string, HighlightLanguage> = {
    cjs: 'javascript',
    js: 'javascript',
    jsx: 'javascript',
    mjs: 'javascript',
    ps1: 'powershell',
    sh: 'bash',
    shell: 'bash',
    ts: 'typescript',
    tsx: 'typescript',
    yml: 'yaml',
    zsh: 'bash'
};

const supportedLanguageSet = new Set<string>(Object.keys(bundledLanguages));

let highlighterPromise: Promise<HighlightHighlighter> | null = null;

function getHighlighter() {
    highlighterPromise ??= createHighlighter({
        themes: [theme],
        langs: Object.keys(bundledLanguages) as HighlightLanguage[]
    });
    return highlighterPromise;
}

function languageFromCodeElement(code: HTMLElement): HighlightLanguage | null {
    const languageClass = Array.from(code.classList).find(className => className.startsWith('language-'));
    const raw = languageClass?.replace('language-', '').trim().toLowerCase();
    if (!raw) {
        return null;
    }
    const aliased = languageAliases[raw] ?? raw;
    return supportedLanguageSet.has(aliased) ? aliased as HighlightLanguage : null;
}

export async function highlightCodeBlocks(root: HTMLElement, signal: AbortSignal) {
    const blocks = Array.from(root.querySelectorAll<HTMLElement>('pre > code'));
    if (blocks.length === 0) {
        return;
    }
    const highlighter = await getHighlighter();
    if (signal.aborted) {
        return;
    }
    for (const code of blocks) {
        if (signal.aborted) {
            return;
        }
        const pre = code.parentElement;
        const lang = languageFromCodeElement(code);
        if (!(pre instanceof HTMLPreElement) || pre.dataset.shiki === 'true' || !lang) {
            continue;
        }
        const html = highlighter.codeToHtml(code.textContent ?? '', {lang, theme});
        const template = document.createElement('template');
        template.innerHTML = html.trim();
        const highlighted = template.content.firstElementChild;
        if (!(highlighted instanceof HTMLPreElement)) {
            continue;
        }
        highlighted.dataset.shiki = 'true';
        highlighted.dataset.language = lang;
        highlighted.classList.add('shiki-block');
        pre.replaceWith(highlighted);
    }
}
