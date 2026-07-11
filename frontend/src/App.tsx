import {useEffect, useMemo, useRef, useState} from 'react';
import {ClipboardSetText, EventsOn, Quit, WindowMinimise, WindowToggleMaximise} from '../wailsjs/runtime/runtime';
import {CheckForUpdates, ConsumeLaunchRequest, DismissUpdate, DownloadUpdate, GetBootstrap, GetReadingMemory, GetReadingPosition, GetReadingSession, LoadDirectory, OpenDocument, OpenDownloadedUpdate, OpenWorkspace, OpenWorkspaceDocument, RefreshWorkspaces, RemoveWorkspace, RenamePath, ReorderWorkspaces, RestoreWorkspaces, SaveOpenTabs, SaveReadingPosition, SearchWorkspace, StatDocument, SwitchLanguage, SwitchTheme, RevealPath} from '../wailsjs/go/main/App';
import type {main} from '../wailsjs/go/models';
import {highlightCodeBlocks} from './codeHighlighter';

type TreeNode = main.TreeNode;
type WorkspaceTree = main.WorkspaceTree;
type WorkspaceCollection = main.WorkspaceCollection;
type Bootstrap = main.Bootstrap;
type Document = main.Document;
type DocumentStatus = main.DocumentStatus;
type Heading = main.Heading;
type ReadingTab = main.ReadingTab;
type ReadingPosition = main.ReadingPosition;
type SearchResult = main.SearchResult;
type UpdateInfo = main.UpdateInfo;
type LaunchRequest = main.LaunchRequest;

const FIND_MATCH_CLASS = 'kern-find-match';
const FIND_CURRENT_CLASS = 'kern-find-current';
const OPEN_DOCS_MIN_HEIGHT = 92;
const WORKSPACE_MIN_HEIGHT = 132;

type DocumentTab = {
    path: string;
    name: string;
};

type OpenDocumentOptions = {
    tabsOverride?: DocumentTab[];
    persistTabs?: boolean;
    savePrevious?: boolean;
    forceReload?: boolean;
    searchTarget?: SearchTarget;
};

type SearchTarget = {
    path: string;
    query: string;
    snippet: string;
};

type ContextMenuState =
    | {kind: 'tree'; node: TreeNode; x: number; y: number}
    | {kind: 'tab'; tab: DocumentTab; x: number; y: number};

type ContextMenuItem = {
    label: string;
    action: () => void | Promise<void>;
    disabled?: boolean;
    separatorBefore?: boolean;
};

type ActionFeedback = {
    kind: 'success' | 'error';
    message: string;
};

function App() {
    const [bootstrap, setBootstrap] = useState<Bootstrap | null>(null);
    const [workspaces, setWorkspaces] = useState<WorkspaceTree[]>([]);
    const [document, setDocument] = useState<Document | null>(null);
    const [tabs, setTabs] = useState<DocumentTab[]>([]);
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
    const [updateInfo, setUpdateInfo] = useState<UpdateInfo | null>(null);
    const [updatePanelOpen, setUpdatePanelOpen] = useState(false);
    const [updateBusy, setUpdateBusy] = useState(false);
    const [updateError, setUpdateError] = useState('');
    const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
    const [renamingPath, setRenamingPath] = useState('');
    const [renameDraft, setRenameDraft] = useState('');
    const [draggingWorkspaceId, setDraggingWorkspaceId] = useState('');
    const [activeHeadingId, setActiveHeadingId] = useState('');
    const [openDocumentsHeight, setOpenDocumentsHeight] = useState(168);
    const [actionFeedback, setActionFeedback] = useState<ActionFeedback | null>(null);
    const draggingWorkspaceIdRef = useRef('');
    const markdownBodyRef = useRef<HTMLDivElement | null>(null);
    const readerScrollRef = useRef<HTMLDivElement | null>(null);
    const readerProgressRef = useRef<HTMLDivElement | null>(null);
    const outlineListRef = useRef<HTMLDivElement | null>(null);
    const findInputRef = useRef<HTMLInputElement | null>(null);
    const searchRequestRef = useRef(0);
    const workspaceRefreshBusyRef = useRef(false);
    const workspaceTreeVersionRef = useRef(0);
    const loadingDirectoryPathsRef = useRef(new Set<string>());
    const readingSaveTimerRef = useRef<number | null>(null);
    const readingRestoreUntilRef = useRef(0);
    const readingNavigationFrameRef = useRef<number | null>(null);
    const readingHeadingsRef = useRef<HTMLElement[] | null>(null);
    const renameCommitBusyRef = useRef(false);
    const actionFeedbackTimerRef = useRef<number | null>(null);
    const searchTargetRef = useRef<SearchTarget | null>(null);
    const documentPathRef = useRef('');
    const shell = bootstrap?.shellLocale ?? {};
    const text = bootstrap?.businessLocale ?? {};
    const product = bootstrap?.product;
    const brand = product?.brandParts;
    const currentTheme = bootstrap?.currentTheme ?? 'system';

    useEffect(() => {
        return () => {
            if (actionFeedbackTimerRef.current !== null) {
                window.clearTimeout(actionFeedbackTimerRef.current);
            }
        };
    }, []);

    useEffect(() => {
        let cancelled = false;
        async function boot() {
            const data = await GetBootstrap('');
            if (cancelled) {
                return;
            }
            setBootstrap(data);
            const restored = await RestoreWorkspaces();
            if (!cancelled && restored?.workspaces?.length) {
                applyWorkspaceCollection(restored);
            }
            const launch = await ConsumeLaunchRequest();
            if (!cancelled && hasLaunchRequest(launch)) {
                await applyLaunchRequest(launch);
                return;
            }
            if (!cancelled && restored?.workspaces?.length) {
                const session = await GetReadingSession();
                const restoredTabs = (session?.openTabs ?? []).map(tabFromReadingSession);
                if (!cancelled && restoredTabs.length) {
                    setTabs(restoredTabs);
                    const activePath = session.activeDocument || restoredTabs[0].path;
                    await openDocument(activePath, session.activePosition ?? null, {
                        tabsOverride: restoredTabs,
                        persistTabs: false,
                        savePrevious: false
                    });
                    return;
                }
                const memory = await GetReadingMemory();
                if (!cancelled && memory?.lastPosition?.path) {
                    await openDocument(memory.lastPosition.path, memory.lastPosition, {
                        persistTabs: false,
                        savePrevious: false
                    });
                }
            }
        }
        boot();
        return () => {
            cancelled = true;
        };
    }, []);

    useEffect(() => {
        if (!bootstrap) {
            return;
        }
        const off = EventsOn('kern:launch-request', () => {
            void ConsumeLaunchRequest()
                .then(request => applyLaunchRequest(request))
                .catch(() => {
                    // Explorer entry points are convenience routes; direct UI actions still surface errors.
                });
        });
        return off;
    }, [bootstrap, document, tabs, documentBusy]);

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
            focusFindInput();
        };
        window.addEventListener('kern:find', openFind);
        return () => {
            window.removeEventListener('kern:find', openFind);
        };
    }, [document]);

    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent) => {
            const key = event.key.toLowerCase();
            const isCommand = event.ctrlKey || event.metaKey;
            if (!isCommand) {
                return;
            }
            if (key === 'a' && document && !isTextInputTarget(event.target)) {
                event.preventDefault();
                selectMarkdownBody();
                return;
            }
            if (key === 'w' && document) {
                event.preventDefault();
                void closeTab(document.path);
                return;
            }
            if (key === 'tab' && tabs.length > 1) {
                event.preventDefault();
                void cycleTab(event.shiftKey ? -1 : 1);
            }
        };
        window.addEventListener('keydown', handleKeyDown, true);
        return () => {
            window.removeEventListener('keydown', handleKeyDown, true);
        };
    }, [document, tabs, documentBusy]);

    useEffect(() => {
        if (!contextMenu) {
            return;
        }
        const closeMenu = () => setContextMenu(null);
        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === 'Escape') {
                closeMenu();
            }
        };
        window.addEventListener('pointerdown', closeMenu);
        window.addEventListener('resize', closeMenu);
        window.addEventListener('scroll', closeMenu, true);
        window.addEventListener('keydown', handleKeyDown, true);
        return () => {
            window.removeEventListener('pointerdown', closeMenu);
            window.removeEventListener('resize', closeMenu);
            window.removeEventListener('scroll', closeMenu, true);
            window.removeEventListener('keydown', handleKeyDown, true);
        };
    }, [contextMenu]);

    useEffect(() => {
        if (findOpen) {
            focusFindInput();
        }
    }, [findOpen]);

    useEffect(() => {
        if (!bootstrap) {
            return;
        }
        let cancelled = false;
        void CheckForUpdates()
            .then(info => {
                if (!cancelled && info?.updateAvailable && !info.ignored) {
                    setUpdateInfo(info);
                }
            })
            .catch(() => {
                // Update checks are intentionally weak and must not block reading.
            });
        return () => {
            cancelled = true;
        };
    }, [bootstrap?.product?.version]);

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
        const reader = readerScrollRef.current;
        if (!reader || !document || documentBusy) {
            return;
        }

        const scheduleSave = () => {
            if (readingSaveTimerRef.current !== null) {
                window.clearTimeout(readingSaveTimerRef.current);
            }
            readingSaveTimerRef.current = window.setTimeout(() => {
                readingSaveTimerRef.current = null;
                void saveCurrentReadingPosition();
            }, 900);
        };

        const idleSaveId = window.setTimeout(() => {
            void saveCurrentReadingPosition();
        }, 1200);
        reader.addEventListener('scroll', scheduleSave, {passive: true});

        return () => {
            window.clearTimeout(idleSaveId);
            if (readingSaveTimerRef.current !== null) {
                window.clearTimeout(readingSaveTimerRef.current);
                readingSaveTimerRef.current = null;
            }
            reader.removeEventListener('scroll', scheduleSave);
        };
    }, [document, documentBusy]);

    useEffect(() => {
        readingHeadingsRef.current = null;
        setActiveHeadingId('');
        if (readerProgressRef.current) {
            readerProgressRef.current.style.transform = 'scaleX(0)';
        }
        const reader = readerScrollRef.current;
        if (!reader || !document || documentBusy) {
            return;
        }
        const scheduleSync = () => scheduleReadingNavigationSync();
        const resizeObserver = new ResizeObserver(scheduleSync);
        reader.addEventListener('scroll', scheduleSync, {passive: true});
        resizeObserver.observe(reader);
        if (markdownBodyRef.current) {
            resizeObserver.observe(markdownBodyRef.current);
        }
        scheduleSync();
        return () => {
            reader.removeEventListener('scroll', scheduleSync);
            resizeObserver.disconnect();
            if (readingNavigationFrameRef.current !== null) {
                window.cancelAnimationFrame(readingNavigationFrameRef.current);
                readingNavigationFrameRef.current = null;
            }
        };
    }, [document?.path, document?.html, documentBusy]);

    useEffect(() => {
        const list = outlineListRef.current;
        if (!list || !activeHeadingId) {
            return;
        }
        const activeRow = Array.from(list.querySelectorAll<HTMLButtonElement>('.outline-row'))
            .find(row => row.dataset.headingId === activeHeadingId);
        if (!activeRow) {
            return;
        }
        const listRect = list.getBoundingClientRect();
        const rowRect = activeRow.getBoundingClientRect();
        const edge = 4;
        if (rowRect.top < listRect.top + edge) {
            list.scrollTop -= listRect.top + edge - rowRect.top;
        } else if (rowRect.bottom > listRect.bottom - edge) {
            list.scrollTop += rowRect.bottom - (listRect.bottom - edge);
        }
    }, [activeHeadingId, document?.path]);

    useEffect(() => {
        if (!workspaces.length) {
            return;
        }
        let cancelled = false;

        const intervalId = window.setInterval(() => {
            void refreshWorkspaceNow(() => !cancelled);
        }, 3000);

        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [workspaces.map(workspace => workspace.workspace.id).join('|')]);

    useEffect(() => {
        const query = searchQuery.trim();
        const requestId = searchRequestRef.current + 1;
        searchRequestRef.current = requestId;
        if (!workspaces.length || Array.from(query).length < 2) {
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
    }, [searchQuery, text, workspaces]);

    function applyWorkspaceCollection(collection: WorkspaceCollection | null | undefined) {
        const nextWorkspaces = collection?.workspaces ?? [];
        workspaceTreeVersionRef.current += 1;
        setWorkspaces(nextWorkspaces);
        setExpandedPaths(current => preserveExpandedWorkspacePaths(current, nextWorkspaces));
    }

    async function applyLaunchRequest(request: LaunchRequest | null | undefined) {
        if (!hasLaunchRequest(request)) {
            return;
        }
        const nextRequest = request;
        if (nextRequest.collection) {
            applyWorkspaceCollection(nextRequest.collection);
        }
        if (nextRequest.documentPath) {
            await openDocument(nextRequest.documentPath);
        }
    }

    async function refreshWorkspaceNow(shouldApply = () => true) {
        if (workspaceRefreshBusyRef.current) {
            return;
        }
        workspaceRefreshBusyRef.current = true;
        const treeVersion = workspaceTreeVersionRef.current;
        try {
            const refresh = await RefreshWorkspaces();
            if (!shouldApply() || treeVersion !== workspaceTreeVersionRef.current || !refresh?.changed || !refresh.collection?.workspaces) {
                return;
            }
            applyWorkspaceCollection(refresh.collection);
            searchRequestRef.current += 1;
            setSearchQuery('');
            setSearchResults([]);
            setSearchBusy(false);
            setSearchError('');
        } catch {
            // Workspace structure refresh is intentionally weak; direct opens surface concrete errors.
        } finally {
            workspaceRefreshBusyRef.current = false;
        }
    }

    async function openWorkspace() {
        if (busy) {
            return;
        }
        await saveCurrentReadingPosition(document, true);
        setBusy(true);
        try {
            const result = await OpenWorkspace();
            if (result?.workspaces?.length) {
                applyWorkspaceCollection(result);
                setSearchQuery('');
                setSearchResults([]);
                setSearchError('');
            }
        } finally {
            setBusy(false);
        }
    }

    async function openDocument(path: string, restorePosition?: ReadingPosition | null, options: OpenDocumentOptions = {}) {
        if (documentBusy) {
            return;
        }
        if (!options.forceReload && document?.path === path) {
            if (options.searchTarget) {
                searchTargetRef.current = options.searchTarget;
                scheduleSearchTargetFocus();
            }
            return;
        }
        searchTargetRef.current = options.searchTarget ?? null;
        const shouldSavePrevious = options.savePrevious ?? Boolean(document?.path && document.path !== path);
        if (shouldSavePrevious) {
            await saveCurrentReadingPosition(document, true);
        }
        setSelectedPath(path);
        setDocumentBusy(true);
        setDocumentError('');
        setStaleStatus(null);
        setDismissedStatusKey('');
        try {
            const result = await OpenDocument(path);
            const position = restorePosition === undefined
                ? await GetReadingPosition(result.path).catch(() => null)
                : restorePosition;
            const nextTabs = upsertTab(options.tabsOverride ?? tabs, documentTabFromDocument(result));
            setTabs(nextTabs);
            if (options.persistTabs !== false) {
                await persistOpenTabs(nextTabs, result.path);
            }
            setDocument(result);
            documentPathRef.current = result.path;
            setSelectedPath(result.path);
            scheduleReaderPosition('', position, result);
            if (options.searchTarget) {
                scheduleSearchTargetFocus();
            }
        } catch (error) {
            searchTargetRef.current = null;
            documentPathRef.current = '';
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
        await saveCurrentReadingPosition(document, true);
        setDocumentBusy(true);
        setDocumentError('');
        setStaleStatus(null);
        setDismissedStatusKey('');
        try {
            const result = await OpenWorkspaceDocument(path);
            const nextTabs = upsertTab(tabs, documentTabFromDocument(result));
            setTabs(nextTabs);
            await persistOpenTabs(nextTabs, result.path);
            setSelectedPath(result.path);
            setDocument(result);
            scheduleReaderPosition(heading, null, result);
        } catch (error) {
            setDocument(null);
            setSelectedPath('');
            setDocumentError(errorMessage(error, text['document.error_unknown']));
        } finally {
            setDocumentBusy(false);
        }
    }

    async function persistOpenTabs(nextTabs: DocumentTab[], activePath: string) {
        await SaveOpenTabs(nextTabs.map(tab => tab.path), activePath).catch(() => {
            // Tab session persistence is a weak convenience; document opens remain the source of visible errors.
        });
    }

    async function closeTab(path: string, event?: React.MouseEvent<HTMLButtonElement>) {
        event?.stopPropagation();
        if (documentBusy) {
            return;
        }
        const tabIndex = tabs.findIndex(tab => tab.path === path);
        if (tabIndex < 0) {
            return;
        }
        const nextTabs = tabs.filter(tab => tab.path !== path);
        const closingActive = document?.path === path;
        if (closingActive) {
            await saveCurrentReadingPosition(document, true);
        }
        if (!nextTabs.length) {
            setTabs([]);
            setDocument(null);
            setSelectedPath('');
            setDocumentError('');
            setStaleStatus(null);
            setDismissedStatusKey('');
            await persistOpenTabs([], '');
            return;
        }
        if (!closingActive) {
            setTabs(nextTabs);
            await persistOpenTabs(nextTabs, document?.path ?? nextTabs[0].path);
            return;
        }
        const nextIndex = Math.min(tabIndex, nextTabs.length - 1);
        const nextTab = nextTabs[nextIndex];
        setTabs(nextTabs);
        await openDocument(nextTab.path, undefined, {
            tabsOverride: nextTabs,
            savePrevious: false
        });
    }

    async function cycleTab(direction: 1 | -1) {
        if (documentBusy || tabs.length < 2) {
            return;
        }
        const currentIndex = Math.max(0, tabs.findIndex(tab => tab.path === document?.path));
        const nextIndex = (currentIndex + direction + tabs.length) % tabs.length;
        await openDocument(tabs[nextIndex].path);
    }

    async function reloadCurrentDocumentPreservingPosition() {
        if (!document || documentBusy) {
            return;
        }
        const position = captureCurrentReadingPosition(document);
        setDocumentBusy(true);
        setDocumentError('');
        setStaleStatus(null);
        setDismissedStatusKey('');
        try {
            const result = await OpenDocument(document.path);
            const restoredPosition = position
                ? {
                    ...position,
                    path: result.path,
                    modifiedAt: result.modifiedAt,
                    size: result.size,
                    updatedAt: Date.now()
                } as ReadingPosition
                : null;
            const nextTabs = upsertTab(tabs, documentTabFromDocument(result));
            setTabs(nextTabs);
            await persistOpenTabs(nextTabs, result.path);
            setDocument(result);
            setSelectedPath(result.path);
            scheduleReaderPosition('', restoredPosition, result);
            if (restoredPosition) {
                window.setTimeout(() => {
                    void SaveReadingPosition(
                        result.path,
                        restoredPosition.scrollTop,
                        restoredPosition.scrollRatio,
                        restoredPosition.headingId,
                        result.modifiedAt,
                        result.size
                    ).catch(() => {
                        // The next scroll event will retry reading-memory persistence.
                    });
                }, 240);
            }
        } catch (error) {
            setDocument(null);
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

    async function toggleDirectory(path: string) {
        if (expandedPaths.has(path)) {
            setExpandedPaths(current => {
                const next = new Set(current);
                next.delete(path);
                return next;
            });
            return;
        }
        const node = findTreeNode(workspaces, path);
        if (!node || loadingDirectoryPathsRef.current.has(path)) {
            return;
        }
        if (!node.loaded) {
            workspaceTreeVersionRef.current += 1;
            loadingDirectoryPathsRef.current.add(path);
            try {
                const loaded = await LoadDirectory(path);
                if (!loaded) {
                    return;
                }
                setWorkspaces(current => replaceTreeNode(current, loaded));
            } catch {
                showActionFeedback(text['feedback.load_directory_failed'], 'error');
                return;
            } finally {
                loadingDirectoryPathsRef.current.delete(path);
            }
        }
        setExpandedPaths(current => new Set(current).add(path));
    }

    function openTreeContextMenu(node: TreeNode, event: React.MouseEvent<HTMLElement>) {
        event.preventDefault();
        event.stopPropagation();
        const {x, y} = contextMenuPoint(event);
        setContextMenu({kind: 'tree', node, x, y});
    }

    function openTabContextMenu(tab: DocumentTab, event: React.MouseEvent<HTMLElement>) {
        event.preventDefault();
        event.stopPropagation();
        const {x, y} = contextMenuPoint(event);
        setContextMenu({kind: 'tab', tab, x, y});
    }

    async function copyPath(path: string) {
        try {
            await ClipboardSetText(path);
            showActionFeedback(text['feedback.copy_success'], 'success');
        } catch {
            showActionFeedback(text['feedback.copy_failed'], 'error');
        }
    }

    async function revealPath(path: string) {
        try {
            await RevealPath(path);
            showActionFeedback(text['feedback.reveal_success'], 'success');
        } catch {
            showActionFeedback(text['feedback.reveal_failed'], 'error');
        }
    }

    function beginRename(node: TreeNode) {
        setRenamingPath(node.path);
        setRenameDraft(node.name);
    }

    function cancelRename() {
        setRenamingPath('');
        setRenameDraft('');
    }

    async function commitRename(node: TreeNode) {
        if (renameCommitBusyRef.current) {
            return;
        }
        const nextName = renameDraft.trim();
        if (!nextName || nextName === node.name) {
            cancelRename();
            return;
        }
        renameCommitBusyRef.current = true;
        workspaceTreeVersionRef.current += 1;
        let renamed = false;
        try {
            const result = await RenamePath(node.path, nextName);
            renamed = true;
            if (result?.tree?.root) {
                setWorkspaces(current => replaceWorkspaceRootForPath(current, result.oldPath, result.tree!.root));
                setExpandedPaths(current => remapExpandedPaths(current, result.oldPath, result.newPath, result.nodeType, workspaces.map(item => item.root)));
            }
            const nextTabs = tabs.map(tab => {
                const path = remapRenamedPath(tab.path, result.oldPath, result.newPath, result.nodeType);
                return path === tab.path ? tab : {path, name: tabNameFromPath(path)};
            });
            const activePath = document?.path ?? '';
            const nextActivePath = activePath ? remapRenamedPath(activePath, result.oldPath, result.newPath, result.nodeType) : '';
            setTabs(nextTabs);
            setSelectedPath(current => current ? remapRenamedPath(current, result.oldPath, result.newPath, result.nodeType) : current);
            await persistOpenTabs(nextTabs, nextActivePath);
            cancelRename();
            if (activePath && nextActivePath !== activePath) {
                const position = captureCurrentReadingPosition(document);
                await openDocument(nextActivePath, position, {
                    tabsOverride: nextTabs,
                    savePrevious: false,
                    forceReload: true
                });
            }
            showActionFeedback(text['feedback.rename_success'], 'success');
        } catch {
            cancelRename();
            showActionFeedback(
                text[renamed ? 'feedback.rename_success' : 'feedback.rename_failed'],
                renamed ? 'success' : 'error'
            );
        } finally {
            renameCommitBusyRef.current = false;
        }
    }

    async function removeWorkspaceRoot(workspace: WorkspaceTree) {
        if (documentBusy) {
            return;
        }
        let removed = false;
        try {
            await saveCurrentReadingPosition(document, true);
            const removedRoot = workspace.root.path;
            const result = await RemoveWorkspace(workspace.workspace.id);
            removed = true;
            applyWorkspaceCollection(result);
            const nextTabs = tabs.filter(tab => !pathIsInsideRoot(tab.path, removedRoot));
            const activeRemoved = document?.path ? pathIsInsideRoot(document.path, removedRoot) : false;
            setTabs(nextTabs);
            setExpandedPaths(current => {
                const next = new Set<string>();
                for (const path of current) {
                    if (!pathIsInsideRoot(path, removedRoot)) {
                        next.add(path);
                    }
                }
                return next;
            });
            if (!activeRemoved) {
                await persistOpenTabs(nextTabs, document?.path ?? '');
                showActionFeedback(text['feedback.remove_workspace_success'], 'success');
                return;
            }
            setDocument(null);
            setSelectedPath('');
            setDocumentError('');
            setStaleStatus(null);
            setDismissedStatusKey('');
            if (nextTabs.length) {
                await openDocument(nextTabs[0].path, undefined, {
                    tabsOverride: nextTabs,
                    savePrevious: false
                });
            } else {
                await persistOpenTabs([], '');
            }
            showActionFeedback(text['feedback.remove_workspace_success'], 'success');
        } catch {
            showActionFeedback(
                text[removed ? 'feedback.remove_workspace_success' : 'feedback.remove_workspace_failed'],
                removed ? 'success' : 'error'
            );
        }
    }

    function showActionFeedback(message: string, kind: ActionFeedback['kind']) {
        if (!message) {
            return;
        }
        if (actionFeedbackTimerRef.current !== null) {
            window.clearTimeout(actionFeedbackTimerRef.current);
        }
        setActionFeedback({kind, message});
        actionFeedbackTimerRef.current = window.setTimeout(() => {
            setActionFeedback(null);
            actionFeedbackTimerRef.current = null;
        }, kind === 'success' ? 2200 : 6000);
    }

    function dismissActionFeedback() {
        if (actionFeedbackTimerRef.current !== null) {
            window.clearTimeout(actionFeedbackTimerRef.current);
            actionFeedbackTimerRef.current = null;
        }
        setActionFeedback(null);
    }

    function beginWorkspaceDrag(workspaceID: string, event: React.DragEvent<HTMLElement>) {
        draggingWorkspaceIdRef.current = workspaceID;
        setDraggingWorkspaceId(workspaceID);
        event.dataTransfer.effectAllowed = 'move';
        event.dataTransfer.setData('text/plain', workspaceID);
    }

    function clearWorkspaceDrag() {
        draggingWorkspaceIdRef.current = '';
        setDraggingWorkspaceId('');
    }

    function allowWorkspaceDrop(event: React.DragEvent<HTMLElement>) {
        if (!draggingWorkspaceIdRef.current) {
            return;
        }
        event.preventDefault();
        event.dataTransfer.dropEffect = 'move';
    }

    async function dropWorkspace(targetWorkspaceID: string, event: React.DragEvent<HTMLElement>) {
        event.preventDefault();
        const sourceWorkspaceID = draggingWorkspaceIdRef.current || event.dataTransfer.getData('text/plain');
        clearWorkspaceDrag();
        if (!sourceWorkspaceID || sourceWorkspaceID === targetWorkspaceID) {
            return;
        }
        const ids = workspaces.map(workspace => workspace.workspace.id);
        const from = ids.indexOf(sourceWorkspaceID);
        const to = ids.indexOf(targetWorkspaceID);
        if (from < 0 || to < 0) {
            return;
        }
        ids.splice(from, 1);
        ids.splice(to, 0, sourceWorkspaceID);
        workspaceTreeVersionRef.current += 1;
        const byID = new Map(workspaces.map(workspace => [workspace.workspace.id, workspace]));
        setWorkspaces(ids.map(id => byID.get(id)).filter(Boolean) as WorkspaceTree[]);
        const result = await ReorderWorkspaces(ids);
        applyWorkspaceCollection(result);
    }

    async function closeOtherTabs(path: string) {
        if (documentBusy) {
            return;
        }
        const targetTab = tabs.find(tab => tab.path === path);
        if (!targetTab) {
            return;
        }
        const nextTabs = [targetTab];
        if (document?.path === path) {
            setTabs(nextTabs);
            await persistOpenTabs(nextTabs, path);
            return;
        }
        await saveCurrentReadingPosition(document, true);
        setTabs(nextTabs);
        await openDocument(path, undefined, {
            tabsOverride: nextTabs,
            savePrevious: false
        });
    }

    async function closeTabsToRight(path: string) {
        if (documentBusy) {
            return;
        }
        const tabIndex = tabs.findIndex(tab => tab.path === path);
        if (tabIndex < 0) {
            return;
        }
        const nextTabs = tabs.slice(0, tabIndex + 1);
        const activePath = document?.path ?? '';
        if (nextTabs.some(tab => tab.path === activePath)) {
            setTabs(nextTabs);
            await persistOpenTabs(nextTabs, activePath || nextTabs[0].path);
            return;
        }
        await saveCurrentReadingPosition(document, true);
        setTabs(nextTabs);
        await openDocument(path, undefined, {
            tabsOverride: nextTabs,
            savePrevious: false
        });
    }

    function contextMenuItems(): ContextMenuItem[] {
        if (!contextMenu) {
            return [];
        }
        if (contextMenu.kind === 'tree') {
            const node = contextMenu.node;
            const isDirectory = node.type === 'directory';
            const isExpanded = expandedPaths.has(node.path);
            const workspace = workspaceForRootPath(workspaces, node.path);
            const isWorkspaceRoot = Boolean(workspace);
            return [
                ...(isDirectory ? [{
                    label: isExpanded ? text['context.collapse'] : text['context.expand'],
                    action: () => toggleDirectory(node.path)
                }] : [{
                    label: text['context.open'],
                    action: () => openDocument(node.path),
                    disabled: documentBusy
                }]),
                {
                    label: text['context.rename'],
                    action: () => beginRename(node),
                    disabled: isWorkspaceRoot || documentBusy
                },
                ...(isWorkspaceRoot && workspace ? [{
                    label: text['context.remove_workspace'],
                    action: () => removeWorkspaceRoot(workspace),
                    separatorBefore: true
                }] : []),
                {
                    label: text['context.refresh_workspace'],
                    action: () => refreshWorkspaceNow(),
                    separatorBefore: !isWorkspaceRoot
                },
                {
                    label: text['context.copy_path'],
                    action: () => copyPath(node.path)
                },
                {
                    label: text['context.show_in_file_manager'],
                    action: () => revealPath(node.path)
                }
            ];
        }

        const tab = contextMenu.tab;
        const tabIndex = tabs.findIndex(item => item.path === tab.path);
        return [
            {
                label: text['context.switch_tab'],
                action: () => openDocument(tab.path),
                disabled: documentBusy || tab.path === document?.path
            },
            {
                label: text['context.close_tab'],
                action: () => closeTab(tab.path),
                disabled: documentBusy
            },
            {
                label: text['context.close_other_tabs'],
                action: () => closeOtherTabs(tab.path),
                disabled: documentBusy || tabs.length <= 1,
                separatorBefore: true
            },
            {
                label: text['context.close_tabs_right'],
                action: () => closeTabsToRight(tab.path),
                disabled: documentBusy || tabIndex < 0 || tabIndex >= tabs.length - 1
            },
            {
                label: text['context.copy_path'],
                action: () => copyPath(tab.path),
                separatorBefore: true
            },
            {
                label: text['context.show_in_file_manager'],
                action: () => revealPath(tab.path)
            }
        ];
    }

    function runContextMenuAction(action: () => void | Promise<void>) {
        setContextMenu(null);
        void Promise.resolve(action()).catch(() => {
            // Context-menu actions are convenience affordances; core document errors stay on direct opens.
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
        const query = searchQuery.trim();
        await openDocument(result.path, null, {
            searchTarget: result.kind === 'content' && query
                ? {path: result.path, query, snippet: result.snippet}
                : undefined
        });
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

    function focusFindInput() {
        window.setTimeout(() => {
            findInputRef.current?.focus();
            findInputRef.current?.select();
        }, 0);
    }

    function selectMarkdownBody() {
        const root = markdownBodyRef.current;
        const selection = window.getSelection();
        if (!root || !selection) {
            return;
        }
        const range = root.ownerDocument.createRange();
        range.selectNodeContents(root);
        selection.removeAllRanges();
        selection.addRange(range);
    }

    function handleTitlebarDoubleClick(event: React.MouseEvent<HTMLElement>) {
        const target = event.target;
        if (target instanceof Element && target.closest('.window-controls')) {
            return;
        }
        void WindowToggleMaximise();
    }

    function runWindowControl(event: React.MouseEvent<HTMLButtonElement>, action: () => void | Promise<void>) {
        event.preventDefault();
        event.stopPropagation();
        void Promise.resolve(action());
    }

    function stopWindowControlPointerEvent(event: React.PointerEvent<HTMLElement> | React.MouseEvent<HTMLElement>) {
        event.stopPropagation();
    }

    function captureCurrentReadingPosition(currentDocument = document): ReadingPosition | null {
        const reader = readerScrollRef.current;
        if (!currentDocument || !reader) {
            return null;
        }
        const maxScroll = Math.max(1, reader.scrollHeight - reader.clientHeight);
        const scrollTop = Math.max(0, Math.round(reader.scrollTop));
        const scrollRatio = Math.min(1, Math.max(0, scrollTop / maxScroll));
        const headingId = currentHeadingId(reader, readingHeadingElements());
        return {
            path: currentDocument.path,
            relativePath: '',
            scrollTop,
            scrollRatio,
            headingId,
            modifiedAt: currentDocument.modifiedAt,
            size: currentDocument.size,
            updatedAt: Date.now()
        } as ReadingPosition;
    }

    async function saveCurrentReadingPosition(currentDocument = document, force = false) {
        if (!force && Date.now() < readingRestoreUntilRef.current) {
            return;
        }
        const position = captureCurrentReadingPosition(currentDocument);
        if (!position || !currentDocument) {
            return;
        }
        await SaveReadingPosition(
            currentDocument.path,
            position.scrollTop,
            position.scrollRatio,
            position.headingId,
            currentDocument.modifiedAt,
            currentDocument.size
        ).catch(() => {
            // Reading memory is a weak convenience; direct document opens still surface real errors.
        });
    }

    function readingHeadingElements() {
        if (readingHeadingsRef.current !== null) {
            return readingHeadingsRef.current;
        }
        const root = markdownBodyRef.current;
        readingHeadingsRef.current = root
            ? Array.from(root.querySelectorAll<HTMLElement>('h1[id], h2[id], h3[id], h4[id], h5[id], h6[id]'))
            : [];
        return readingHeadingsRef.current;
    }

    function scheduleReadingNavigationSync() {
        if (readingNavigationFrameRef.current !== null) {
            return;
        }
        readingNavigationFrameRef.current = window.requestAnimationFrame(() => {
            readingNavigationFrameRef.current = null;
            const reader = readerScrollRef.current;
            const root = markdownBodyRef.current;
            if (!reader || !root) {
                setActiveHeadingId('');
                if (readerProgressRef.current) {
                    readerProgressRef.current.style.transform = 'scaleX(0)';
                }
                return;
            }
            const maxScroll = Math.max(0, reader.scrollHeight - reader.clientHeight);
            const progress = maxScroll === 0 ? 1 : Math.min(1, Math.max(0, reader.scrollTop / maxScroll));
            if (readerProgressRef.current) {
                readerProgressRef.current.style.transform = `scaleX(${progress})`;
            }
            const headings = readingHeadingElements();
            const nextHeadingId = currentHeadingId(reader, headings) || headings[0]?.id || '';
            setActiveHeadingId(current => current === nextHeadingId ? current : nextHeadingId);
        });
    }

    async function downloadUpdate() {
        if (!updateInfo || updateBusy) {
            return;
        }
        setUpdateBusy(true);
        setUpdateError('');
        try {
            const downloaded = await DownloadUpdate(updateInfo.downloadUrl, updateInfo.sha256);
            setUpdateInfo({
                ...updateInfo,
                downloadedPath: downloaded.downloadedPath,
                sha256: downloaded.sha256 || updateInfo.sha256
            } as UpdateInfo);
        } catch (error) {
            setUpdateError(errorMessage(error, text['update.error']));
        } finally {
            setUpdateBusy(false);
        }
    }

    async function installUpdate() {
        if (!updateInfo?.downloadedPath || updateBusy) {
            return;
        }
        setUpdateBusy(true);
        setUpdateError('');
        try {
            await OpenDownloadedUpdate(updateInfo.downloadedPath);
        } catch (error) {
            setUpdateError(errorMessage(error, text['update.error']));
        } finally {
            setUpdateBusy(false);
        }
    }

    async function dismissUpdate() {
        if (!updateInfo || updateBusy) {
            return;
        }
        setUpdateBusy(true);
        setUpdateError('');
        try {
            await DismissUpdate(updateInfo.latestVersion);
            setUpdateInfo(null);
            setUpdatePanelOpen(false);
        } catch (error) {
            setUpdateError(errorMessage(error, text['update.error']));
        } finally {
            setUpdateBusy(false);
        }
    }

    function scheduleReaderPosition(heading = '', position: ReadingPosition | null = null, currentDocument: Document | null = document) {
        const restoreUntil = Date.now() + 900;
        readingRestoreUntilRef.current = restoreUntil;
        window.requestAnimationFrame(() => {
            window.requestAnimationFrame(() => {
                setReaderPosition(heading, position, currentDocument);
                window.setTimeout(() => {
                    if (readingRestoreUntilRef.current === restoreUntil) {
                        readingRestoreUntilRef.current = 0;
                    }
                }, 120);
            });
        });
    }

    function setReaderPosition(heading = '', position: ReadingPosition | null = null, currentDocument: Document | null = document) {
        const reader = readerScrollRef.current;
        if (!reader) {
            return;
        }
        const positionMatchesDocument = position && currentDocument &&
            position.modifiedAt === currentDocument.modifiedAt &&
            position.size === currentDocument.size;
        if (positionMatchesDocument) {
            const maxScroll = Math.max(0, reader.scrollHeight - reader.clientHeight);
            const ratioTop = Math.round(maxScroll * Math.min(1, Math.max(0, position.scrollRatio)));
            const targetTop = position.scrollTop > maxScroll ? ratioTop : position.scrollTop;
            reader.scrollTop = Math.max(0, Math.min(maxScroll, targetTop));
            reader.scrollLeft = 0;
            scheduleReadingNavigationSync();
            return;
        }
        if (!heading && position?.headingId) {
            heading = position.headingId;
        }
        if (heading) {
            const target = globalThis.document.getElementById(heading);
            if (target) {
                const targetTop = target.getBoundingClientRect().top - reader.getBoundingClientRect().top + reader.scrollTop;
                reader.scrollTop = Math.max(0, targetTop);
                reader.scrollLeft = 0;
                scheduleReadingNavigationSync();
                return;
            }
        }
        reader.scrollTop = 0;
        reader.scrollLeft = 0;
        scheduleReadingNavigationSync();
    }

    function scheduleSearchTargetFocus() {
        window.requestAnimationFrame(() => {
            window.requestAnimationFrame(() => {
                const target = searchTargetRef.current;
                const root = markdownBodyRef.current;
                const reader = readerScrollRef.current;
                if (!target || !root || !reader || documentPathRef.current !== target.path) {
                    return;
                }
                clearFindMarks(root);
                const matches = applyFindMarks(root, target.query);
                const match = searchTargetMatch(root, matches, target.snippet) ?? matches[0];
                if (!match) {
                    return;
                }
                match.classList.add(FIND_CURRENT_CLASS);
                const targetTop = match.getBoundingClientRect().top - reader.getBoundingClientRect().top + reader.scrollTop - 80;
                reader.scrollTop = Math.max(0, targetTop);
                scheduleReadingNavigationSync();
            });
        });
    }

    function beginSidebarResize(event: React.PointerEvent<HTMLDivElement>) {
        const sidebar = event.currentTarget.closest<HTMLElement>('.sidebar');
        if (!sidebar) {
            return;
        }
        event.preventDefault();
        const rect = sidebar.getBoundingClientRect();
        const maxHeight = Math.max(OPEN_DOCS_MIN_HEIGHT, rect.height - WORKSPACE_MIN_HEIGHT);
        const handlePointerMove = (moveEvent: PointerEvent) => {
            const nextHeight = Math.min(maxHeight, Math.max(OPEN_DOCS_MIN_HEIGHT, moveEvent.clientY - rect.top));
            setOpenDocumentsHeight(Math.round(nextHeight));
        };
        const stopResize = () => {
            globalThis.document.body.classList.remove('resizing-sidebar');
            window.removeEventListener('pointermove', handlePointerMove);
            window.removeEventListener('pointerup', stopResize);
            window.removeEventListener('pointercancel', stopResize);
        };
        globalThis.document.body.classList.add('resizing-sidebar');
        window.addEventListener('pointermove', handlePointerMove);
        window.addEventListener('pointerup', stopResize);
        window.addEventListener('pointercancel', stopResize);
    }

    const hasWorkspace = useMemo(() => workspaces.length > 0, [workspaces]);
    const showSearchPanel = searchQuery.trim().length >= 2 && hasWorkspace;

    if (!bootstrap) {
        return <div className="boot-frame"/>;
    }

    return (
        <main className="app-shell">
            <header className="titlebar" onDoubleClick={handleTitlebarDoubleClick}>
                <div className="brand-lockup" aria-label={product?.displayName}>
                    <span className="brand-prefix">{brand?.prefix}</span>
                    <span className="brand-core">{brand?.core}</span>
                    <span className="brand-suffix">{brand?.suffix}</span>
                </div>
                <div className="titlebar-center">
                    <span className="subtitle">{text['app.subtitle']}</span>
                </div>
                <div className="window-controls" onPointerDown={stopWindowControlPointerEvent} onMouseDown={stopWindowControlPointerEvent} onDoubleClick={event => event.stopPropagation()}>
                    <button type="button" title={shell['window.minimize']} aria-label={shell['window.minimize']} onClick={event => runWindowControl(event, WindowMinimise)}>-</button>
                    <button type="button" title={shell['window.maximize']} aria-label={shell['window.maximize']} onClick={event => runWindowControl(event, WindowToggleMaximise)}>□</button>
                    <button type="button" className="close" title={shell['window.close']} aria-label={shell['window.close']} onClick={event => runWindowControl(event, Quit)}>×</button>
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
                                                {result.matchLine ? <span className="search-result-line">{formatLineLabel(text['search.line'], result.matchLine)}</span> : null}
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
                {updateInfo?.updateAvailable ? (
                    <div className="update-area">
                        <button type="button" className="update-button" onClick={() => setUpdatePanelOpen(current => !current)}>
                            {text['update.available']} {updateInfo.latestVersion}
                        </button>
                        {updatePanelOpen ? (
                            <div className="update-popover">
                                <div className="update-title">{text['update.available']}</div>
                                <div className="update-version">{text['update.version']} {updateInfo.latestVersion}</div>
                                {updateInfo.releaseNotes ? (
                                    <div className="update-notes">
                                        <div className="update-notes-label">{text['update.notes']}</div>
                                        <div className="update-notes-body">{updateInfo.releaseNotes}</div>
                                    </div>
                                ) : null}
                                {updateError ? <div className="update-error">{updateError}</div> : null}
                                <div className="update-actions">
                                    {updateInfo.downloadedPath ? (
                                        <button type="button" onClick={installUpdate} disabled={updateBusy}>
                                            {text['update.install']}
                                        </button>
                                    ) : (
                                        <button type="button" onClick={downloadUpdate} disabled={updateBusy || !updateInfo.downloadUrl}>
                                            {updateBusy ? text['update.downloading'] : text['update.download']}
                                        </button>
                                    )}
                                    <button type="button" onClick={dismissUpdate} disabled={updateBusy}>
                                        {text['update.dismiss']}
                                    </button>
                                </div>
                            </div>
                        ) : null}
                    </div>
                ) : null}
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
                <aside className="sidebar" style={{'--open-documents-h': `${openDocumentsHeight}px`} as React.CSSProperties}>
                    <section className="sidebar-section open-documents-section">
                        <div className="panel-heading">{text['panel.open_documents']}</div>
                        <div className="open-documents-scroll">
                            {tabs.length ? (
                                <div className="open-documents-list" role="list" aria-label={text['tabs.label']}>
                                    {tabs.map(tab => {
                                        const active = tab.path === document?.path;
                                        const workspace = workspaceForDocumentPath(workspaces, tab.path);
                                        const external = !workspace;
                                        return (
                                            <div
                                                className={`open-document-row ${active ? 'active' : ''}`}
                                                key={tab.path}
                                                role="listitem"
                                                onContextMenu={event => openTabContextMenu(tab, event)}
                                            >
                                                <button
                                                    type="button"
                                                    className="open-document-main"
                                                    title={tab.path}
                                                    onClick={() => openDocument(tab.path)}
                                                >
                                                    <span className="open-document-name">{tab.name}</span>
                                                    <span className="open-document-meta">
                                                        {external ? text['tabs.external'] : workspace.workspace.name}
                                                    </span>
                                                </button>
                                                <button
                                                    type="button"
                                                    className="open-document-close"
                                                    title={text['tabs.close']}
                                                    aria-label={`${text['tabs.close']} ${tab.name}`}
                                                    onClick={event => closeTab(tab.path, event)}
                                                >
                                                    ×
                                                </button>
                                            </div>
                                        );
                                    })}
                                </div>
                            ) : (
                                <div className="empty-state compact">{text['empty.open_documents']}</div>
                            )}
                        </div>
                    </section>
                    <div className="sidebar-splitter" role="separator" aria-orientation="horizontal" onPointerDown={beginSidebarResize}/>
                    <section className="sidebar-section workspace-section">
                        <div className="panel-heading">{text['panel.workspace']}</div>
                        <div className="tree-scroll">
                            {hasWorkspace ? (
                                <div className="workspace-roots">
                                    {workspaces.map(workspace => (
                                        <TreeView
                                            key={workspace.workspace.id}
                                            node={workspace.root}
                                            depth={0}
                                            workspaceId={workspace.workspace.id}
                                            isWorkspaceRoot
                                            selectedPath={selectedPath}
                                            expandedPaths={expandedPaths}
                                            renamingPath={renamingPath}
                                            renameDraft={renameDraft}
                                            onRenameDraftChange={setRenameDraft}
                                            onCommitRename={commitRename}
                                            onCancelRename={cancelRename}
                                            onToggleDirectory={toggleDirectory}
                                            onOpenDocument={openDocument}
                                            onOpenContextMenu={openTreeContextMenu}
                                            onWorkspaceDragStart={beginWorkspaceDrag}
                                            onWorkspaceDragOver={allowWorkspaceDrop}
                                            onWorkspaceDrop={dropWorkspace}
                                            onWorkspaceDragEnd={clearWorkspaceDrag}
                                            draggingWorkspaceId={draggingWorkspaceId}
                                        />
                                    ))}
                                </div>
                            ) : (
                                <div className="empty-state">{text['empty.workspace']}</div>
                            )}
                        </div>
                    </section>
                </aside>

                <article className="reader-surface">
                    {tabs.length ? (
                        <div className="document-tabs" role="tablist" aria-label={text['tabs.label']}>
                            {tabs.map(tab => {
                                const active = tab.path === document?.path;
                                return (
                                    <div
                                        className={`document-tab ${active ? 'active' : ''}`}
                                        key={tab.path}
                                        onContextMenu={event => openTabContextMenu(tab, event)}
                                    >
                                        <button
                                            type="button"
                                            className="document-tab-main"
                                            role="tab"
                                            aria-selected={active}
                                            title={tab.path}
                                            onClick={() => openDocument(tab.path)}
                                        >
                                            <span className="document-tab-name">{tab.name}</span>
                                        </button>
                                        <button
                                            type="button"
                                            className="document-tab-close"
                                            title={text['tabs.close']}
                                            aria-label={`${text['tabs.close']} ${tab.name}`}
                                            onClick={event => closeTab(tab.path, event)}
                                        >
                                            ×
                                        </button>
                                    </div>
                                );
                            })}
                        </div>
                    ) : null}
                    {document ? (
                        <div className="reader-progress-track" aria-hidden="true">
                            <div className="reader-progress-value" ref={readerProgressRef}/>
                        </div>
                    ) : null}
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
                    {(staleStatus && document && !documentError) || actionFeedback ? (
                        <div className="reader-notice-stack">
                            {staleStatus && document && !documentError ? (
                                <div className="reader-status-bar">
                                    <div className="document-message changed-message">
                                        <div>
                                            <div className="message-title">{text['document.changed_title']}</div>
                                            <div className="message-body">{text['document.changed_body']}</div>
                                        </div>
                                        <div className="message-actions">
                                            <button type="button" onClick={reloadCurrentDocumentPreservingPosition} disabled={documentBusy}>
                                                {text['document.reload']}
                                            </button>
                                            <button type="button" onClick={dismissDocumentChange}>
                                                {text['document.dismiss']}
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ) : null}
                            {actionFeedback ? (
                                <div className="reader-feedback-bar">
                                    <div
                                        className={`action-feedback ${actionFeedback.kind}`}
                                        role={actionFeedback.kind === 'error' ? 'alert' : 'status'}
                                        aria-live={actionFeedback.kind === 'error' ? 'assertive' : 'polite'}
                                    >
                                        <span className="action-feedback-message">{actionFeedback.message}</span>
                                        <button
                                            type="button"
                                            className="action-feedback-close"
                                            title={text['feedback.dismiss']}
                                            aria-label={text['feedback.dismiss']}
                                            onClick={dismissActionFeedback}
                                        >
                                            ×
                                        </button>
                                    </div>
                                </div>
                            ) : null}
                        </div>
                    ) : null}
                </article>

                <aside className="outline-panel">
                    <div className="panel-heading">{text['panel.outline']}</div>
                    {document?.outline?.length ? (
                        <div className="outline-list" ref={outlineListRef}>
                            {document.outline.map((heading, index) => (
                                <button
                                    type="button"
                                    key={`${heading.id}-${index}`}
                                    className={`outline-row ${heading.id === activeHeadingId ? 'active' : ''}`}
                                    style={{'--level': heading.level} as React.CSSProperties}
                                    data-heading-id={heading.id}
                                    aria-current={heading.id === activeHeadingId ? 'location' : undefined}
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
            {contextMenu ? (
                <ContextMenu
                    x={contextMenu.x}
                    y={contextMenu.y}
                    items={contextMenuItems()}
                    onAction={runContextMenuAction}
                />
            ) : null}
        </main>
    );
}

function tabFromReadingSession(tab: ReadingTab): DocumentTab {
    return {
        path: tab.path,
        name: tab.name || tabNameFromPath(tab.path)
    };
}

function documentTabFromDocument(currentDocument: Document): DocumentTab {
    return {
        path: currentDocument.path,
        name: currentDocument.name || tabNameFromPath(currentDocument.path)
    };
}

function upsertTab(currentTabs: DocumentTab[], nextTab: DocumentTab) {
    if (!nextTab.path) {
        return currentTabs;
    }
    const index = currentTabs.findIndex(tab => tab.path === nextTab.path);
    if (index < 0) {
        return [...currentTabs, nextTab];
    }
    const nextTabs = [...currentTabs];
    nextTabs[index] = nextTab;
    return nextTabs;
}

function tabNameFromPath(path: string) {
    return path.split(/[\\/]/).filter(Boolean).pop() || path;
}

function documentStatusKey(status: DocumentStatus) {
    return `${status.exists}:${status.isDocument}:${status.modifiedAt}:${status.size}`;
}

function currentHeadingId(reader: HTMLElement, headings: HTMLElement[]) {
    if (!headings.length) {
        return '';
    }
    const readerTop = reader.getBoundingClientRect().top;
    let current = '';
    for (const heading of headings) {
        const offset = heading.getBoundingClientRect().top - readerTop;
        if (offset <= 96) {
            current = heading.id;
        } else {
            break;
        }
    }
    return current;
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

function isTextInputTarget(target: EventTarget | null) {
    return target instanceof Element && Boolean(target.closest('input, textarea, [contenteditable="true"]'));
}

function hasLaunchRequest(request: LaunchRequest | null | undefined): request is LaunchRequest {
    return Boolean(request?.collection?.workspaces?.length || request?.documentPath);
}

function workspaceForRootPath(workspaces: WorkspaceTree[], path: string) {
    return workspaces.find(workspace => normalizePathForCompare(workspace.root.path) === normalizePathForCompare(path));
}

function workspaceForDocumentPath(workspaces: WorkspaceTree[], path: string) {
    return workspaces.find(workspace => pathIsInsideRoot(path, workspace.root.path));
}

function replaceWorkspaceRootForPath(workspaces: WorkspaceTree[], path: string, root: TreeNode) {
    return workspaces.map(workspace => pathIsInsideRoot(path, workspace.root.path)
        ? ({...workspace, root} as WorkspaceTree)
        : workspace);
}

function findTreeNode(workspaces: WorkspaceTree[], path: string) {
    for (const workspace of workspaces) {
        const found = findNode(workspace.root, path);
        if (found) {
            return found;
        }
    }
    return null;
}

function findNode(node: TreeNode, path: string): TreeNode | null {
    if (normalizePathForCompare(node.path) === normalizePathForCompare(path)) {
        return node;
    }
    for (const child of node.children ?? []) {
        const found = findNode(child, path);
        if (found) {
            return found;
        }
    }
    return null;
}

function replaceTreeNode(workspaces: WorkspaceTree[], replacement: TreeNode) {
    return workspaces.map(workspace => ({
        ...workspace,
        root: replaceNode(workspace.root, replacement)
    } as WorkspaceTree));
}

function replaceNode(node: TreeNode, replacement: TreeNode): TreeNode {
    if (normalizePathForCompare(node.path) === normalizePathForCompare(replacement.path)) {
        return replacement;
    }
    if (!node.children?.length || !pathIsInsideRoot(replacement.path, node.path)) {
        return node;
    }
    return {
        ...node,
        children: node.children.map(child => replaceNode(child, replacement))
    } as TreeNode;
}

function pathIsInsideRoot(path: string, root: string) {
    const normalizedPath = normalizePathForCompare(path);
    const normalizedRoot = normalizePathForCompare(root);
    return normalizedPath === normalizedRoot || normalizedPath.startsWith(normalizedRoot + '/');
}

function collectDirectoryPaths(node: TreeNode, paths = new Set<string>()) {
    if (node.type !== 'directory') {
        return paths;
    }
    paths.add(node.path);
    for (const child of node.children ?? []) {
        collectDirectoryPaths(child, paths);
    }
    return paths;
}

function preserveExpandedWorkspacePaths(current: Set<string>, workspaces: WorkspaceTree[]) {
    const directoryPaths = new Set<string>();
    for (const workspace of workspaces) {
        collectDirectoryPaths(workspace.root, directoryPaths);
    }
    const next = new Set<string>();
    for (const path of current) {
        if (directoryPaths.has(path)) {
            next.add(path);
        }
    }
    return next;
}

function remapExpandedPaths(current: Set<string>, oldPath: string, newPath: string, nodeType: string, roots: TreeNode[]) {
    const mapped = new Set<string>();
    for (const path of current) {
        mapped.add(remapRenamedPath(path, oldPath, newPath, nodeType));
    }
    const directoryPaths = new Set<string>();
    for (const root of roots) {
        collectDirectoryPaths(root, directoryPaths);
    }
    const next = new Set<string>();
    for (const path of mapped) {
        if (directoryPaths.has(path)) {
            next.add(path);
        }
    }
    return next;
}

function remapRenamedPath(path: string, oldPath: string, newPath: string, nodeType: string) {
    const slashPath = pathWithForwardSlashes(path);
    const slashOld = pathWithForwardSlashes(oldPath);
    const normalizedPath = normalizePathForCompare(path);
    const normalizedOld = normalizePathForCompare(oldPath);
    if (normalizedPath === normalizedOld) {
        return newPath;
    }
    if (nodeType === 'directory' && normalizedPath.startsWith(normalizedOld + '/')) {
        const suffix = slashPath.slice(slashOld.length);
        const separator = newPath.includes('\\') ? '\\' : '/';
        return newPath + suffix.replace(/\//g, separator);
    }
    return path;
}

function pathWithForwardSlashes(path: string) {
    return path.replace(/\\/g, '/').replace(/\/+$/g, '');
}

function normalizePathForCompare(path: string) {
    return pathWithForwardSlashes(path).toLocaleLowerCase();
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

function searchTargetMatch(root: HTMLElement, matches: HTMLElement[], snippet: string) {
    const compactSnippet = compactSearchText(snippet);
    if (!compactSnippet) {
        return null;
    }
    return matches.find(match => {
        let element = match.parentElement;
        while (element && element !== root) {
            if (compactSearchText(element.textContent ?? '').includes(compactSnippet)) {
                return true;
            }
            element = element.parentElement;
        }
        return false;
    }) ?? null;
}

function compactSearchText(value: string) {
    return value.replace(/\s+/g, ' ').trim().toLocaleLowerCase();
}

function formatLineLabel(template: string | undefined, line: number) {
    return (template || 'Line {line}').replace('{line}', String(line));
}

function updateCurrentFindMatch(matches: HTMLElement[], index: number) {
    matches.forEach((match, matchIndex) => {
        match.classList.toggle(FIND_CURRENT_CLASS, matchIndex === index);
    });
}

function contextMenuPoint(event: React.MouseEvent<HTMLElement>) {
    const width = 220;
    const height = 260;
    const edge = 8;
    return {
        x: Math.max(edge, Math.min(event.clientX, window.innerWidth - width - edge)),
        y: Math.max(edge, Math.min(event.clientY, window.innerHeight - height - edge))
    };
}

function ContextMenu({
    x,
    y,
    items,
    onAction
}: {
    x: number;
    y: number;
    items: ContextMenuItem[];
    onAction: (action: () => void | Promise<void>) => void;
}) {
    return (
        <div
            className="context-menu"
            role="menu"
            style={{left: x, top: y} as React.CSSProperties}
            onPointerDown={event => event.stopPropagation()}
            onContextMenu={event => event.preventDefault()}
        >
            {items.map((item, index) => (
                <button
                    type="button"
                    role="menuitem"
                    key={`${item.label}-${index}`}
                    className={`context-menu-item ${item.separatorBefore ? 'with-separator' : ''}`}
                    disabled={item.disabled}
                    onClick={() => onAction(item.action)}
                >
                    {item.label}
                </button>
            ))}
        </div>
    );
}

function TreeView({
    node,
    depth,
    workspaceId,
    isWorkspaceRoot = false,
    selectedPath,
    expandedPaths,
    renamingPath,
    renameDraft,
    onRenameDraftChange,
    onCommitRename,
    onCancelRename,
    onToggleDirectory,
    onOpenDocument,
    onOpenContextMenu,
    onWorkspaceDragStart,
    onWorkspaceDragOver,
    onWorkspaceDrop,
    onWorkspaceDragEnd,
    draggingWorkspaceId
}: {
    node: TreeNode;
    depth: number;
    workspaceId?: string;
    isWorkspaceRoot?: boolean;
    selectedPath: string;
    expandedPaths: Set<string>;
    renamingPath: string;
    renameDraft: string;
    onRenameDraftChange: (value: string) => void;
    onCommitRename: (node: TreeNode) => void;
    onCancelRename: () => void;
    onToggleDirectory: (path: string) => void | Promise<void>;
    onOpenDocument: (path: string) => void;
    onOpenContextMenu: (node: TreeNode, event: React.MouseEvent<HTMLElement>) => void;
    onWorkspaceDragStart?: (workspaceID: string, event: React.DragEvent<HTMLElement>) => void;
    onWorkspaceDragOver?: (event: React.DragEvent<HTMLElement>) => void;
    onWorkspaceDrop?: (workspaceID: string, event: React.DragEvent<HTMLElement>) => void;
    onWorkspaceDragEnd?: () => void;
    draggingWorkspaceId?: string;
}) {
    const isDirectory = node.type === 'directory';
    const isSelected = node.path === selectedPath;
    const isExpanded = isDirectory && expandedPaths.has(node.path);
    const isRenaming = renamingPath === node.path;
    return (
        <div className="tree-node" style={{'--depth': depth} as React.CSSProperties}>
            {isRenaming ? (
                <div className={`tree-row rename-row ${isDirectory ? 'directory' : 'file'} ${isSelected ? 'selected' : ''}`}>
                    <span className="tree-glyph">{isDirectory ? (isExpanded ? '▾' : '▸') : '•'}</span>
                    <input
                        className="tree-rename-input"
                        value={renameDraft}
                        autoFocus
                        autoComplete="off"
                        autoCorrect="off"
                        autoCapitalize="off"
                        spellCheck={false}
                        aria-label={node.name}
                        onChange={event => onRenameDraftChange(event.currentTarget.value)}
                        onBlur={event => {
                            if (event.currentTarget.dataset.cancelled === 'true') {
                                return;
                            }
                            onCommitRename(node);
                        }}
                        onClick={event => event.stopPropagation()}
                        onKeyDown={event => {
                            if (event.key === 'Enter') {
                                event.preventDefault();
                                onCommitRename(node);
                            }
                            if (event.key === 'Escape') {
                                event.preventDefault();
                                event.currentTarget.dataset.cancelled = 'true';
                                onCancelRename();
                            }
                        }}
                    />
                </div>
            ) : (
                <button
                    type="button"
                    className={`tree-row ${isDirectory ? 'directory' : 'file'} ${isWorkspaceRoot ? 'workspace-root' : ''} ${isWorkspaceRoot && workspaceId === draggingWorkspaceId ? 'dragging' : ''} ${isSelected ? 'selected' : ''}`}
                    draggable={isWorkspaceRoot}
                    data-kern-draggable-workspace={isWorkspaceRoot ? 'true' : undefined}
                    onDragStart={event => {
                        if (isWorkspaceRoot && workspaceId) {
                            onWorkspaceDragStart?.(workspaceId, event);
                        }
                    }}
                    onDragOver={event => {
                        if (isWorkspaceRoot) {
                            onWorkspaceDragOver?.(event);
                        }
                    }}
                    onDrop={event => {
                        if (isWorkspaceRoot && workspaceId) {
                            onWorkspaceDrop?.(workspaceId, event);
                        }
                    }}
                    onDragEnd={() => {
                        if (isWorkspaceRoot) {
                            onWorkspaceDragEnd?.();
                        }
                    }}
                    onClick={() => isDirectory ? onToggleDirectory(node.path) : onOpenDocument(node.path)}
                    onContextMenu={event => onOpenContextMenu(node, event)}
                    aria-expanded={isDirectory ? isExpanded : undefined}
                >
                    <span className="tree-glyph">{isDirectory ? (isExpanded ? '▾' : '▸') : '•'}</span>
                    <span className="tree-name" title={node.path}>{node.name}</span>
                </button>
            )}
            {isExpanded && node.children?.length ? (
                <div className="tree-children">
                    {node.children.map(child => (
                        <TreeView
                            key={child.path}
                            node={child}
                            depth={depth + 1}
                            workspaceId={workspaceId}
                            selectedPath={selectedPath}
                            expandedPaths={expandedPaths}
                            renamingPath={renamingPath}
                            renameDraft={renameDraft}
                            onRenameDraftChange={onRenameDraftChange}
                            onCommitRename={onCommitRename}
                            onCancelRename={onCancelRename}
                            onToggleDirectory={onToggleDirectory}
                            onOpenDocument={onOpenDocument}
                            onOpenContextMenu={onOpenContextMenu}
                            onWorkspaceDragStart={onWorkspaceDragStart}
                            onWorkspaceDragOver={onWorkspaceDragOver}
                            onWorkspaceDrop={onWorkspaceDrop}
                            onWorkspaceDragEnd={onWorkspaceDragEnd}
                            draggingWorkspaceId={draggingWorkspaceId}
                        />
                    ))}
                </div>
            ) : null}
        </div>
    );
}

export default App;
