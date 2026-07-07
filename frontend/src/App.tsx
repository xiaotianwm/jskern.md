import {useEffect, useMemo, useRef, useState} from 'react';
import {Quit, WindowMinimise, WindowToggleMaximise} from '../wailsjs/runtime/runtime';
import {GetBootstrap, OpenDocument, OpenWorkspace, OpenWorkspaceDocument, RestoreWorkspace, SearchWorkspace, StatDocument, SwitchLanguage, SwitchTheme} from '../wailsjs/go/main/App';
import type {main} from '../wailsjs/go/models';
import {highlightCodeBlocks} from './codeHighlighter';

type TreeNode = main.TreeNode;
type Bootstrap = main.Bootstrap;
type Document = main.Document;
type DocumentStatus = main.DocumentStatus;
type Heading = main.Heading;
type SearchResult = main.SearchResult;

const FIND_MATCH_CLASS = 'kern-find-match';
const FIND_CURRENT_CLASS = 'kern-find-current';

function App() {
    const [bootstrap, setBootstrap] = useState<Bootstrap | null>(null);
    const [tree, setTree] = useState<TreeNode | null>(null);
    const [document, setDocument] = useState<Document | null>(null);
    const [selectedPath, setSelectedPath] = useState<string>('');
    const [expandedPaths, setExpandedPaths] = useState<Set<string>>(new Set());
    const [busy, setBusy] = useState(false);
    const [documentBusy, setDocumentBusy] = useState(false);
    const [documentError, setDocumentError] = useState('');
    const [staleStatus, setStaleStatus] = useState<DocumentStatus | null>(null);
    const [dismissedStatusKey, setDismissedStatusKey] = useState('');
    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
    const [searchBusy, setSearchBusy] = useState(false);
    const [searchError, setSearchError] = useState('');
    const [findOpen, setFindOpen] = useState(false);
    const [findQuery, setFindQuery] = useState('');
    const [findMatches, setFindMatches] = useState<HTMLElement[]>([]);
    const [findIndex, setFindIndex] = useState(0);
    const markdownBodyRef = useRef<HTMLDivElement | null>(null);
    const readerScrollRef = useRef<HTMLDivElement | null>(null);
    const findInputRef = useRef<HTMLInputElement | null>(null);
    const searchRequestRef = useRef(0);
    const shell = bootstrap?.shellLocale ?? {};
    const text = bootstrap?.businessLocale ?? {};
    const product = bootstrap?.product;
    const brand = product?.brandParts;
    const currentTheme = bootstrap?.currentTheme ?? 'system';

    useEffect(() => {
        let cancelled = false;
        async function boot() {
            const data = await GetBootstrap('');
            if (cancelled) {
                return;
            }
            setBootstrap(data);
            const restored = await RestoreWorkspace();
            if (!cancelled && restored?.root) {
                setTree(restored.root);
                setExpandedPaths(new Set([restored.root.path]));
            }
        }
        boot();
        return () => {
            cancelled = true;
        };
    }, []);

    useEffect(() => {
        const media = window.matchMedia('(prefers-color-scheme: dark)');
        const applyTheme = () => {
            const useDark = currentTheme === 'dark' || (currentTheme === 'system' && media.matches);
            globalThis.document.documentElement.classList.toggle('dark', useDark);
        };
        applyTheme();
        media.addEventListener('change', applyTheme);
        return () => {
            media.removeEventListener('change', applyTheme);
        };
    }, [currentTheme]);

    useEffect(() => {
        const root = markdownBodyRef.current;
        if (!root || !document?.html) {
            return;
        }
        const controller = new AbortController();
        void highlightCodeBlocks(root, controller.signal).catch(() => {
            // Keep Go-rendered plain code blocks if Shiki cannot initialize.
        });
        return () => {
            controller.abort();
        };
    }, [document?.html]);

    useEffect(() => {
        const openFind = () => {
            if (!document) {
                return;
            }
            setFindOpen(true);
            window.requestAnimationFrame(() => {
                findInputRef.current?.focus();
                findInputRef.current?.select();
            });
        };
        window.addEventListener('kern:find', openFind);
        return () => {
            window.removeEventListener('kern:find', openFind);
        };
    }, [document]);

    useEffect(() => {
        setFindOpen(false);
        setFindQuery('');
        setFindMatches([]);
        setFindIndex(0);
        const root = markdownBodyRef.current;
        if (root) {
            clearFindMarks(root);
        }
    }, [document?.path]);

    useEffect(() => {
        const root = markdownBodyRef.current;
        if (!root) {
            setFindMatches([]);
            setFindIndex(0);
            return;
        }
        clearFindMarks(root);
        const query = findQuery.trim();
        if (!query) {
            setFindMatches([]);
            setFindIndex(0);
            return;
        }
        const matches = applyFindMarks(root, query);
        setFindMatches(matches);
        setFindIndex(0);
    }, [document?.html, findQuery]);

    useEffect(() => {
        updateCurrentFindMatch(findMatches, findIndex);
        const current = findMatches[findIndex];
        const reader = readerScrollRef.current;
        if (!current || !reader) {
            return;
        }
        const targetTop = current.getBoundingClientRect().top - reader.getBoundingClientRect().top + reader.scrollTop - 80;
        reader.scrollTop = Math.max(0, targetTop);
    }, [findIndex, findMatches]);

    useEffect(() => {
        if (!document || documentBusy) {
            return;
        }
        const currentDocument = document;
        let cancelled = false;

        async function checkDocumentStatus() {
            try {
                const status = await StatDocument(currentDocument.path, currentDocument.modifiedAt, currentDocument.size);
                if (cancelled) {
                    return;
                }
                const statusKey = documentStatusKey(status);
                if (status.changed && statusKey !== dismissedStatusKey) {
                    setStaleStatus(status);
                    return;
                }
                if (!status.changed) {
                    setStaleStatus(null);
                    setDismissedStatusKey('');
                }
            } catch {
                // Keep this reminder weak; document reload will surface concrete errors.
            }
        }

        const intervalId = window.setInterval(() => {
            void checkDocumentStatus();
        }, 5000);

        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [dismissedStatusKey, document, documentBusy]);

    useEffect(() => {
        const query = searchQuery.trim();
        const requestId = searchRequestRef.current + 1;
        searchRequestRef.current = requestId;
        if (!tree || Array.from(query).length < 2) {
            setSearchResults([]);
            setSearchBusy(false);
            setSearchError('');
            return;
        }

        const timeoutId = window.setTimeout(() => {
            setSearchBusy(true);
            setSearchError('');
            void SearchWorkspace(query)
                .then(results => {
                    if (searchRequestRef.current === requestId) {
                        setSearchResults(results ?? []);
                    }
                })
                .catch(error => {
                    if (searchRequestRef.current === requestId) {
                        setSearchResults([]);
                        setSearchError(errorMessage(error, text['search.error']));
                    }
                })
                .finally(() => {
                    if (searchRequestRef.current === requestId) {
                        setSearchBusy(false);
                    }
                });
        }, 180);

        return () => {
            window.clearTimeout(timeoutId);
        };
    }, [searchQuery, text, tree]);

    async function openWorkspace() {
        if (busy) {
            return;
        }
        setBusy(true);
        try {
            const result = await OpenWorkspace();
            if (result?.root) {
                setTree(result.root);
                setDocument(null);
                setSelectedPath('');
                setDocumentError('');
                setStaleStatus(null);
                setDismissedStatusKey('');
                setSearchQuery('');
                setSearchResults([]);
                setSearchError('');
                setExpandedPaths(new Set([result.root.path]));
            }
        } finally {
            setBusy(false);
        }
    }

    async function openDocument(path: string) {
        if (documentBusy) {
            return;
        }
        setSelectedPath(path);
        setDocumentBusy(true);
        setDocumentError('');
        setStaleStatus(null);
        setDismissedStatusKey('');
        try {
            const result = await OpenDocument(path);
            setDocument(result);
            setSelectedPath(result.path);
            scheduleReaderPosition();
        } catch (error) {
            setDocument(null);
            setDocumentError(errorMessage(error, text['document.error_unknown']));
        } finally {
            setDocumentBusy(false);
        }
    }

    async function openLinkedDocument(path: string, heading: string) {
        if (documentBusy) {
            return;
        }
        setDocumentBusy(true);
        setDocumentError('');
        setStaleStatus(null);
        setDismissedStatusKey('');
        try {
            const result = await OpenWorkspaceDocument(path);
            setSelectedPath(result.path);
            setDocument(result);
            scheduleReaderPosition(heading);
        } catch (error) {
            setDocument(null);
            setSelectedPath('');
            setDocumentError(errorMessage(error, text['document.error_unknown']));
        } finally {
            setDocumentBusy(false);
        }
    }

    function handleMarkdownClick(event: React.MouseEvent<HTMLDivElement>) {
        const target = event.target;
        if (!(target instanceof Element)) {
            return;
        }
        const link = target.closest<HTMLAnchorElement>('a[data-kern-document]');
        if (!link) {
            return;
        }
        event.preventDefault();
        const path = link.dataset.kernDocument ?? '';
        const heading = link.dataset.kernHeading ?? '';
        if (path) {
            void openLinkedDocument(path, heading);
        }
    }

    function scrollToHeading(heading: Heading) {
        if (!heading.id) {
            return;
        }
        setReaderPosition(heading.id);
    }

    function toggleDirectory(path: string) {
        setExpandedPaths(current => {
            const next = new Set(current);
            if (next.has(path)) {
                next.delete(path);
            } else {
                next.add(path);
            }
            return next;
        });
    }

    function dismissDocumentChange() {
        if (!staleStatus) {
            return;
        }
        setDismissedStatusKey(documentStatusKey(staleStatus));
        setStaleStatus(null);
    }

    async function openSearchResult(result: SearchResult) {
        await openDocument(result.path);
        setSearchQuery('');
        setSearchResults([]);
        setSearchError('');
    }

    async function switchLanguage(locale: string) {
        const data = await SwitchLanguage(locale);
        setBootstrap(data);
    }

    async function switchTheme(theme: string) {
        const data = await SwitchTheme(theme);
        setBootstrap(data);
    }

    function handleSearchKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
        if (event.key === 'Enter' && searchResults.length > 0) {
            event.preventDefault();
            void openSearchResult(searchResults[0]);
        }
        if (event.key === 'Escape') {
            setSearchQuery('');
            setSearchResults([]);
            setSearchError('');
        }
    }

    function closeFind() {
        setFindOpen(false);
        setFindQuery('');
        setFindMatches([]);
        setFindIndex(0);
        const root = markdownBodyRef.current;
        if (root) {
            clearFindMarks(root);
        }
    }

    function moveFindMatch(direction: 1 | -1) {
        if (!findMatches.length) {
            return;
        }
        setFindIndex(current => (current + direction + findMatches.length) % findMatches.length);
    }

    function handleFindKeyDown(event: React.KeyboardEvent<HTMLInputElement>) {
        if (event.key === 'Enter') {
            event.preventDefault();
            moveFindMatch(event.shiftKey ? -1 : 1);
        }
        if (event.key === 'Escape') {
            event.preventDefault();
            closeFind();
        }
    }

    function scheduleReaderPosition(heading = '') {
        window.requestAnimationFrame(() => {
            window.requestAnimationFrame(() => {
                setReaderPosition(heading);
            });
        });
    }

    function setReaderPosition(heading = '') {
        const reader = readerScrollRef.current;
        if (!reader) {
            return;
        }
        if (heading) {
            const target = globalThis.document.getElementById(heading);
            if (target) {
                const targetTop = target.getBoundingClientRect().top - reader.getBoundingClientRect().top + reader.scrollTop;
                reader.scrollTop = Math.max(0, targetTop);
                reader.scrollLeft = 0;
                return;
            }
        }
        reader.scrollTop = 0;
        reader.scrollLeft = 0;
    }

    const hasWorkspace = useMemo(() => Boolean(tree), [tree]);
    const showSearchPanel = searchQuery.trim().length >= 2 && hasWorkspace;

    if (!bootstrap) {
        return <div className="boot-frame"/>;
    }

    return (
        <main className="app-shell">
            <header className="titlebar">
                <div className="brand-lockup" aria-label={product?.displayName}>
                    <span className="brand-prefix">{brand?.prefix}</span>
                    <span className="brand-core">{brand?.core}</span>
                    <span className="brand-suffix">{brand?.suffix}</span>
                </div>
                <div className="titlebar-center">
                    <span className="subtitle">{text['app.subtitle']}</span>
                </div>
                <div className="window-controls">
                    <button type="button" title={shell['window.minimize']} aria-label={shell['window.minimize']} onClick={WindowMinimise}>-</button>
                    <button type="button" title={shell['window.maximize']} aria-label={shell['window.maximize']} onClick={WindowToggleMaximise}>□</button>
                    <button type="button" className="close" title={shell['window.close']} aria-label={shell['window.close']} onClick={Quit}>×</button>
                </div>
            </header>

            <section className="toolbar">
                <button type="button" className="primary-action" onClick={openWorkspace} disabled={busy}>
                    {text['action.open_workspace']}
                </button>
                <div className="search-box">
                    <input
                        className="search-input"
                        type="text"
                        value={searchQuery}
                        placeholder={text['search.placeholder']}
                        aria-label={text['search.placeholder']}
                        disabled={!hasWorkspace}
                        autoComplete="off"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                        onChange={event => setSearchQuery(event.currentTarget.value)}
                        onKeyDown={handleSearchKeyDown}
                    />
                    {showSearchPanel ? (
                        <div className="search-results">
                            {searchBusy ? (
                                <div className="search-state">{text['search.loading']}</div>
                            ) : searchError ? (
                                <div className="search-state error-text">{searchError}</div>
                            ) : searchResults.length ? (
                                searchResults.map((result, index) => (
                                    <button
                                        type="button"
                                        key={`${result.kind}-${result.path}-${index}`}
                                        className="search-result-row"
                                        onClick={() => openSearchResult(result)}
                                    >
                                        <span className="search-result-kind">{text[`search.${result.kind}`]}</span>
                                        <span className="search-result-main">
                                            <span className="search-result-name">{result.name}</span>
                                            <span className="search-result-path">{result.relativePath}</span>
                                            {result.snippet ? <span className="search-result-snippet">{result.snippet}</span> : null}
                                        </span>
                                    </button>
                                ))
                            ) : (
                                <div className="search-state">{text['search.empty']}</div>
                            )}
                        </div>
                    ) : null}
                </div>
                <span className="status-line">{text['status.reader_first']}</span>
                <div className="toolbar-spacer"/>
                <label className="toolbar-select">
                    <span>{shell['menu.language']}</span>
                    <select
                        value={bootstrap.currentLocale}
                        aria-label={shell['menu.language']}
                        onChange={event => void switchLanguage(event.currentTarget.value)}
                    >
                        {product?.languages?.map(language => (
                            <option key={language.code} value={language.code}>{language.label}</option>
                        ))}
                    </select>
                </label>
                <label className="toolbar-select">
                    <span>{shell['menu.theme']}</span>
                    <select
                        value={currentTheme}
                        aria-label={shell['menu.theme']}
                        onChange={event => void switchTheme(event.currentTarget.value)}
                    >
                        <option value="system">{shell['theme.system']}</option>
                        <option value="light">{shell['theme.light']}</option>
                        <option value="dark">{shell['theme.dark']}</option>
                    </select>
                </label>
            </section>

            <section className="workspace-layout">
                <aside className="sidebar">
                    <div className="panel-heading">{text['panel.workspace']}</div>
                    <div className="tree-scroll">
                        {hasWorkspace && tree ? (
                            <TreeView
                                node={tree}
                                depth={0}
                                selectedPath={selectedPath}
                                expandedPaths={expandedPaths}
                                onToggleDirectory={toggleDirectory}
                                onOpenDocument={openDocument}
                            />
                        ) : (
                            <div className="empty-state">{text['empty.workspace']}</div>
                        )}
                    </div>
                </aside>

                <article className="reader-surface">
                    {findOpen && document ? (
                        <div className="find-bar">
                            <input
                                ref={findInputRef}
                                className="find-input"
                                type="text"
                                value={findQuery}
                                placeholder={text['find.placeholder']}
                                aria-label={text['find.placeholder']}
                                autoComplete="off"
                                autoCorrect="off"
                                autoCapitalize="off"
                                spellCheck={false}
                                onChange={event => setFindQuery(event.currentTarget.value)}
                                onKeyDown={handleFindKeyDown}
                            />
                            <span className={`find-counter ${findQuery.trim() && !findMatches.length ? 'empty' : ''}`}>
                                {findQuery.trim() && !findMatches.length ? text['find.no_matches'] : `${findMatches.length ? findIndex + 1 : 0}/${findMatches.length}`}
                            </span>
                            <button type="button" className="find-nav" title={text['find.previous']} aria-label={text['find.previous']} onClick={() => moveFindMatch(-1)} disabled={!findMatches.length}>
                                ↑
                            </button>
                            <button type="button" className="find-nav" title={text['find.next']} aria-label={text['find.next']} onClick={() => moveFindMatch(1)} disabled={!findMatches.length}>
                                ↓
                            </button>
                            <button type="button" className="find-close" title={text['find.close']} aria-label={text['find.close']} onClick={closeFind}>
                                ×
                            </button>
                        </div>
                    ) : null}
                    <div className="reader-scroll" ref={readerScrollRef}>
                        {documentError ? (
                            <div className="document-message error-message">
                                <div className="message-title">{text['document.error_title']}</div>
                                <div className="message-body">
                                    {text['document.error_open_failed']} <span className="selectable-data">{documentError}</span>
                                </div>
                            </div>
                        ) : document ? (
                            <div className="document-view">
                                <header className="document-header">
                                    <h1>{document.title}</h1>
                                    <div className="document-path selectable-data">{document.path}</div>
                                </header>
                                <div
                                    ref={markdownBodyRef}
                                    className="markdown-body"
                                    onClick={handleMarkdownClick}
                                    dangerouslySetInnerHTML={{__html: document.html}}
                                />
                            </div>
                        ) : (
                            <div className="reader-placeholder" aria-busy={documentBusy}>
                                <div className="document-mark"/>
                                <div className="line wide"/>
                                <div className="line"/>
                                <div className="line short"/>
                            </div>
                        )}
                    </div>
                    {staleStatus && document && !documentError ? (
                        <div className="reader-status-bar">
                            <div className="document-message changed-message">
                                <div>
                                    <div className="message-title">{text['document.changed_title']}</div>
                                    <div className="message-body">{text['document.changed_body']}</div>
                                </div>
                                <div className="message-actions">
                                    <button type="button" onClick={() => openDocument(document.path)} disabled={documentBusy}>
                                        {text['document.reload']}
                                    </button>
                                    <button type="button" onClick={dismissDocumentChange}>
                                        {text['document.dismiss']}
                                    </button>
                                </div>
                            </div>
                        </div>
                    ) : null}
                </article>

                <aside className="outline-panel">
                    <div className="panel-heading">{text['panel.outline']}</div>
                    {document?.outline?.length ? (
                        <div className="outline-list">
                            {document.outline.map((heading, index) => (
                                <button
                                    type="button"
                                    key={`${heading.id}-${index}`}
                                    className="outline-row"
                                    style={{'--level': heading.level} as React.CSSProperties}
                                    onClick={() => scrollToHeading(heading)}
                                >
                                    {heading.text}
                                </button>
                            ))}
                        </div>
                    ) : (
                        <div className="empty-state">{text['empty.outline']}</div>
                    )}
                </aside>
            </section>
        </main>
    );
}

function documentStatusKey(status: DocumentStatus) {
    return `${status.exists}:${status.isDocument}:${status.modifiedAt}:${status.size}`;
}

function errorMessage(error: unknown, fallback: string) {
    if (error instanceof Error && error.message) {
        return error.message;
    }
    if (typeof error === 'string' && error) {
        return error;
    }
    return fallback;
}

function clearFindMarks(root: HTMLElement) {
    const marks = Array.from(root.querySelectorAll(`mark.${FIND_MATCH_CLASS}`));
    for (const mark of marks) {
        const parent = mark.parentNode;
        if (!parent) {
            continue;
        }
        parent.replaceChild(globalThis.document.createTextNode(mark.textContent ?? ''), mark);
        parent.normalize();
    }
}

function applyFindMarks(root: HTMLElement, query: string) {
    const needle = query.toLocaleLowerCase();
    const textNodes: Text[] = [];
    const walker = globalThis.document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
        acceptNode(node) {
            const parent = node.parentElement;
            if (!parent || parent.closest(`mark.${FIND_MATCH_CLASS}, script, style, input, textarea`)) {
                return NodeFilter.FILTER_REJECT;
            }
            return (node.nodeValue ?? '').toLocaleLowerCase().includes(needle)
                ? NodeFilter.FILTER_ACCEPT
                : NodeFilter.FILTER_REJECT;
        }
    });
    while (walker.nextNode()) {
        textNodes.push(walker.currentNode as Text);
    }

    const marks: HTMLElement[] = [];
    for (const node of textNodes) {
        const value = node.nodeValue ?? '';
        const lower = value.toLocaleLowerCase();
        const fragment = globalThis.document.createDocumentFragment();
        let cursor = 0;
        let matchIndex = lower.indexOf(needle);
        while (matchIndex >= 0) {
            if (matchIndex > cursor) {
                fragment.append(value.slice(cursor, matchIndex));
            }
            const mark = globalThis.document.createElement('mark');
            mark.className = FIND_MATCH_CLASS;
            mark.textContent = value.slice(matchIndex, matchIndex + query.length);
            fragment.append(mark);
            marks.push(mark);
            cursor = matchIndex + query.length;
            matchIndex = lower.indexOf(needle, cursor);
        }
        if (cursor < value.length) {
            fragment.append(value.slice(cursor));
        }
        node.replaceWith(fragment);
    }
    return marks;
}

function updateCurrentFindMatch(matches: HTMLElement[], index: number) {
    matches.forEach((match, matchIndex) => {
        match.classList.toggle(FIND_CURRENT_CLASS, matchIndex === index);
    });
}

function TreeView({
    node,
    depth,
    selectedPath,
    expandedPaths,
    onToggleDirectory,
    onOpenDocument
}: {
    node: TreeNode;
    depth: number;
    selectedPath: string;
    expandedPaths: Set<string>;
    onToggleDirectory: (path: string) => void;
    onOpenDocument: (path: string) => void;
}) {
    const isDirectory = node.type === 'directory';
    const isSelected = node.path === selectedPath;
    const isExpanded = isDirectory && expandedPaths.has(node.path);
    return (
        <div className="tree-node" style={{'--depth': depth} as React.CSSProperties}>
            <button
                type="button"
                className={`tree-row ${isDirectory ? 'directory' : 'file'} ${isSelected ? 'selected' : ''}`}
                onClick={() => isDirectory ? onToggleDirectory(node.path) : onOpenDocument(node.path)}
                aria-expanded={isDirectory ? isExpanded : undefined}
            >
                <span className="tree-glyph">{isDirectory ? (isExpanded ? '▾' : '▸') : '•'}</span>
                <span className="tree-name" title={node.path}>{node.name}</span>
            </button>
            {isExpanded && node.children?.length ? (
                <div className="tree-children">
                    {node.children.map(child => (
                        <TreeView
                            key={child.path}
                            node={child}
                            depth={depth + 1}
                            selectedPath={selectedPath}
                            expandedPaths={expandedPaths}
                            onToggleDirectory={onToggleDirectory}
                            onOpenDocument={onOpenDocument}
                        />
                    ))}
                </div>
            ) : null}
        </div>
    );
}

export default App;
