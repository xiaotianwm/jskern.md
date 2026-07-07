export function installDesktopGuards() {
    window.addEventListener('contextmenu', event => event.preventDefault());
    window.addEventListener('dragstart', event => event.preventDefault());
    window.addEventListener('wheel', event => {
        if (event.ctrlKey) {
            event.preventDefault();
        }
    }, {passive: false});
    window.addEventListener('keydown', event => {
        const key = event.key.toLowerCase();
        const blocked =
            key === 'f5' ||
            key === 'f12' ||
            ((event.ctrlKey || event.metaKey) && ['r', 'f', '+', '-', '0'].includes(key));
        if (blocked) {
            event.preventDefault();
            event.stopPropagation();
        }
    });
}
