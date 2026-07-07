import {useEffect, useMemo, useState} from 'react';
import {Quit, WindowMinimise, WindowToggleMaximise} from '../wailsjs/runtime/runtime';
import {GetBootstrap, OpenWorkspace} from '../wailsjs/go/main/App';
import type {main} from '../wailsjs/go/models';

type TreeNode = main.TreeNode;
type Bootstrap = main.Bootstrap;

function App() {
    const [bootstrap, setBootstrap] = useState<Bootstrap | null>(null);
    const [tree, setTree] = useState<TreeNode | null>(null);
    const [busy, setBusy] = useState(false);

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
            }
        } finally {
            setBusy(false);
        }
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
                        <TreeView node={tree} depth={0}/>
                    ) : (
                        <div className="empty-state">{text['empty.workspace']}</div>
                    )}
                </aside>

                <article className="reader-surface">
                    <div className="reader-placeholder">
                        <div className="document-mark"/>
                        <div className="line wide"/>
                        <div className="line"/>
                        <div className="line short"/>
                    </div>
                </article>

                <aside className="outline-panel">
                    <div className="panel-heading">{text['panel.outline']}</div>
                    <div className="empty-state">{text['empty.outline']}</div>
                </aside>
            </section>
        </main>
    );
}

function TreeView({node, depth}: { node: TreeNode; depth: number }) {
    const isDirectory = node.type === 'directory';
    return (
        <div className="tree-node" style={{'--depth': depth} as React.CSSProperties}>
            <div className={`tree-row ${isDirectory ? 'directory' : 'file'}`}>
                <span className="tree-glyph">{isDirectory ? '▸' : '•'}</span>
                <span className="tree-name" title={node.path}>{node.name}</span>
            </div>
            {isDirectory && node.children?.length ? (
                <div className="tree-children">
                    {node.children.map(child => (
                        <TreeView key={child.path} node={child} depth={depth + 1}/>
                    ))}
                </div>
            ) : null}
        </div>
    );
}

export default App;
