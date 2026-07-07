import {useEffect, useMemo, useState} from 'react';
import {Quit, WindowMinimise, WindowToggleMaximise} from '../wailsjs/runtime/runtime';
import {GetBootstrap, OpenDocument, OpenWorkspace} from '../wailsjs/go/main/App';
import type {main} from '../wailsjs/go/models';

type TreeNode = main.TreeNode;
type Bootstrap = main.Bootstrap;
type Document = main.Document;
type Heading = main.Heading;

function App() {
    const [bootstrap, setBootstrap] = useState<Bootstrap | null>(null);
    const [tree, setTree] = useState<TreeNode | null>(null);
    const [document, setDocument] = useState<Document | null>(null);
    const [selectedPath, setSelectedPath] = useState<string>('');
    const [busy, setBusy] = useState(false);
    const [documentBusy, setDocumentBusy] = useState(false);

    useEffect(() => {
        GetBootstrap('zh-CN').then(setBootstrap);
    }, []);

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
        try {
            const result = await OpenDocument(path);
            setDocument(result);
        } finally {
            setDocumentBusy(false);
        }
    }

    function scrollToHeading(heading: Heading) {
        if (!heading.id) {
            return;
        }
        globalThis.document.getElementById(heading.id)?.scrollIntoView({block: 'start'});
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
                    {hasWorkspace && tree ? (
                        <TreeView node={tree} depth={0} selectedPath={selectedPath} onOpenDocument={openDocument}/>
                    ) : (
                        <div className="empty-state">{text['empty.workspace']}</div>
                    )}
                </aside>

                <article className="reader-surface">
                    {document ? (
                        <div className="document-view">
                            <header className="document-header">
                                <h1>{document.title}</h1>
                                <div className="document-path selectable-data">{document.path}</div>
                            </header>
                            <div className="markdown-body" dangerouslySetInnerHTML={{__html: document.html}}/>
                        </div>
                    ) : (
                        <div className="reader-placeholder" aria-busy={documentBusy}>
                            <div className="document-mark"/>
                            <div className="line wide"/>
                            <div className="line"/>
                            <div className="line short"/>
                        </div>
                    )}
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

function TreeView({
    node,
    depth,
    selectedPath,
    onOpenDocument
}: {
    node: TreeNode;
    depth: number;
    selectedPath: string;
    onOpenDocument: (path: string) => void;
}) {
    const isDirectory = node.type === 'directory';
    const isSelected = node.path === selectedPath;
    return (
        <div className="tree-node" style={{'--depth': depth} as React.CSSProperties}>
            <button
                type="button"
                className={`tree-row ${isDirectory ? 'directory' : 'file'} ${isSelected ? 'selected' : ''}`}
                onClick={() => !isDirectory && onOpenDocument(node.path)}
                disabled={isDirectory}
            >
                <span className="tree-glyph">{isDirectory ? '▸' : '•'}</span>
                <span className="tree-name" title={node.path}>{node.name}</span>
            </button>
            {isDirectory && node.children?.length ? (
                <div className="tree-children">
                    {node.children.map(child => (
                        <TreeView
                            key={child.path}
                            node={child}
                            depth={depth + 1}
                            selectedPath={selectedPath}
                            onOpenDocument={onOpenDocument}
                        />
                    ))}
                </div>
            ) : null}
        </div>
    );
}

export default App;
