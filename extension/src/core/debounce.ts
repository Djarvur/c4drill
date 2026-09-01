// Debounce helper for the live preview (issue #27): coalesces bursts of
// document edits into a single trailing re-render. Pure logic — no VS Code
// imports — so it is unit-testable with node:test.

export interface Debounced<A extends unknown[]> {
    (...args: A): void;

    /** Cancels a pending invocation without running it. */
    cancel(): void;

    /** Runs a pending invocation immediately, if any. */
    flush(): void;
}

// debounce returns a trailing-edge debounced wrapper around fn. The wrapped
// function exposes cancel/flush for lifecycle cleanup (panel disposal).
export function debounce<A extends unknown[]>(fn: (...args: A) => void, waitMs: number): Debounced<A> {
    let timer: NodeJS.Timeout | undefined;
    let lastArgs: A | undefined;

    const wrapped = (...args: A): void => {
        lastArgs = args;

        if (timer !== undefined) {
            clearTimeout(timer);
        }

        timer = setTimeout(() => {
            timer = undefined;

            const run = lastArgs;
            lastArgs = undefined;
            if (run !== undefined) {
                fn(...run);
            }
        }, waitMs);
    };

    wrapped.cancel = (): void => {
        if (timer !== undefined) {
            clearTimeout(timer);
            timer = undefined;
        }

        lastArgs = undefined;
    };

    wrapped.flush = (): void => {
        if (timer === undefined) {
            return;
        }

        clearTimeout(timer);
        timer = undefined;

        const run = lastArgs;
        lastArgs = undefined;
        if (run !== undefined) {
            fn(...run);
        }
    };

    return wrapped;
}
