import {useEffect, useMemo, useRef, useState} from 'react';
import {Quit, WindowMinimise, WindowToggleMaximise} from '../wailsjs/runtime/runtime';
import {GetBootstrap, OpenDocument, OpenWorkspace, OpenWorkspaceDocument, RestoreWorkspace, StatDocument} from '../wailsjs/go/main/App';
import type {main} from '../wailsjs/go/models';
import {highlightCodeBlocks} from './codeHighlighter';

type TreeNode = main.TreeNode;
type Bootstrap = main.Bootstrap;
type Document = main.Document;
type DocumentStatus = main.DocumentStatus;
type Heading = main.Heading;

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
    const markdownBodyRef = useRef<HTMLDivElement | null>(null);
    const readerScrollRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        let cancelled = false;
        async function boot() {
            const data = await GetBootstrap('zh-CN');
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

    const shell = bootstrap?.shellLocale ?? {};
    const text = bootstrap?.businessLocale ?? {};
    const product = bootstrap?.product;
    const brand = product?.brandParts;

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
                <span className="status-line">{text['status.reader_first']}</span>
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
