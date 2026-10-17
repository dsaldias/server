// components/baseui/floating_focus_manager.js
// Port of @base-ui/react floating-ui-react/components/FloatingFocusManager.tsx
// (1.6.0), with utils/FocusGuard.tsx and floating-ui-react/utils/enqueueFocus.ts.
//
// The component's effects become one handle that follows the popup:
//
//   const focus = window.templ.focusManager.useFloatingFocusManager(options)  on mount
//   focus.close(details)   when it closes, details { reason, event }
//   focus.unmount()        once it unmounted after its exit animation
//
//   floating                  the floating element (positioner, or the popup itself)
//   reference                 the trigger, the source's domReference
//   triggers                  every trigger of the popup (a list), default [reference]
//   modal                     traps focus with guards and hides the outside (default true)
//   initialFocus              true, false, an element, or fn(interactionType)
//   returnFocus               true, false, an element, or fn(closeType)
//   restoreFocus              false, true, or "popup"
//   closeOnFocusOut           default true
//   openInteractionType       "mouse", "touch", "pen", "keyboard", "", or null (programmatic)
//   previousFocusableElement, nextFocusableElement
//   getInsideElements         fn returning elements that count as inside
//   onOpenChange(open, reason, event)   requests the close
//
// The floating tree is the portal owner chain, like in use_dismiss.js: a
// child is an open focus manager whose floating element lies inside this one.
// The portal context is the [data-base-ui-portal] node around the floating
// element, see portal.js.
(function () {
  "use strict";

  const t = () => window.templ.tabbable;
  const webkit = typeof CSS !== "undefined" && !!CSS.supports?.("-webkit-backdrop-filter:none");
  const FOCUSABLE_ATTRIBUTE = "data-base-ui-focusable";
  const CLICK_TRIGGER_IDENTIFIER = "data-base-ui-click-trigger";
  const TYPEABLE_SELECTOR = "input:not([type='hidden']):not([disabled]),[contenteditable]:not([contenteditable='false']),textarea:not([disabled])";

  const instances = new Set();

  // ----- enqueueFocus -------------------------------------------------------

  // Returns a cancel for the queued focus.
  let rafId = 0;
  function enqueueFocus(el, options = {}) {
    const { preventScroll = false, sync = false, shouldFocus } = options;
    cancelAnimationFrame(rafId);
    const exec = () => {
      if (shouldFocus && !shouldFocus()) return;
      el?.focus({ preventScroll });
    };
    if (sync) {
      exec();
      return () => {};
    }
    const currentRafId = requestAnimationFrame(exec);
    rafId = currentRafId;
    return () => {
      if (rafId === currentRafId) {
        cancelAnimationFrame(currentRafId);
        rafId = 0;
      }
    };
  }

  // ----- FocusGuard -----------------------------------------------------------

  // A visually hidden tabbable span that hands focus on when it receives it.
  // @base-ui/utils/platform: VoiceOver may run on any Apple OS. Through
  // WebKit its virtual cursor only focuses focusable or button elements, so
  // the guards there are buttons the cursor can land on instead of hidden.
  const platform = (navigator.platform || "").toLowerCase();
  const ios = /^i(os$|p)/.test(platform) || (platform === "macintel" && navigator.maxTouchPoints > 1);
  const apple = ios || platform.startsWith("mac");
  const voiceOverGuards = apple && webkit;

  function createFocusGuard(type, onFocus) {
    const guard = document.createElement("span");
    if (type) guard.setAttribute("data-type", type);
    if (voiceOverGuards) guard.setAttribute("role", "button");
    else guard.setAttribute("aria-hidden", "true");
    guard.setAttribute("tabindex", "0");
    guard.setAttribute("data-base-ui-focus-guard", "");
    guard.style.cssText = "clip-path:inset(50%);overflow:hidden;white-space:nowrap;border:0;padding:0;width:1px;height:1px;margin:-1px;position:fixed;top:0;left:0";
    if (onFocus) guard.addEventListener("focus", onFocus);
    return guard;
  }

  // ----- helpers from the source's utils --------------------------------------

  const getTarget = (event) => event.composedPath?.()[0] || event.target;
  const isHTMLElement = (node) => node instanceof HTMLElement;
  const isTypeableElement = (element) => isHTMLElement(element) && element.matches(TYPEABLE_SELECTOR);
  const isTypeableCombobox = (element) => !!element && element.getAttribute("role") === "combobox" && isTypeableElement(element);
  const stopEvent = (event) => {
    event.preventDefault();
    event.stopPropagation();
  };

  // The element that takes focus: the one marked focusable inside a
  // positioning wrapper, or the floating element itself.
  function getFloatingFocusElement(floatingElement) {
    if (!floatingElement) return null;
    return floatingElement.hasAttribute(FOCUSABLE_ATTRIBUTE)
      ? floatingElement
      : floatingElement.querySelector(`[${FOCUSABLE_ATTRIBUTE}]`) || floatingElement;
  }

  function getFirstTabbableElement(container) {
    if (!container) return null;
    if (t().isTabbable(container)) return container;
    return t().tabbable(container)[0] || container;
  }

  // The React tree pendant, see use_dismiss.js.
  function withinTree(root, target) {
    for (let node = target; node; node = window.templ.portal.treeParent(node)) {
      if (node === root) return true;
    }
    return false;
  }

  function isVirtualClick(event) {
    if (event.pointerType === "" && event.isTrusted) return true;
    return event.detail === 0 && !event.pointerType;
  }

  function isVirtualPointerEvent(event) {
    return (event.width === 0 && event.height === 0) ||
      (event.width < 1 && event.height < 1 && event.pressure === 0 && event.detail === 0 && event.pointerType === "touch");
  }

  function getEventType(event, lastInteractionType) {
    if (!event) return lastInteractionType || "";
    if (event instanceof KeyboardEvent) return "keyboard";
    if (event instanceof FocusEvent) return lastInteractionType || "keyboard";
    if ("pointerType" in event) return event.pointerType || "keyboard";
    if ("touches" in event) return "touch";
    if (event instanceof MouseEvent) return lastInteractionType || (event.detail === 0 ? "keyboard" : "mouse");
    return "";
  }

  // The elements focused before a modal popup took focus, most recent last.
  const LIST_LIMIT = 20;
  let previouslyFocusedElements = [];
  function clearDisconnectedPreviouslyFocusedElements() {
    previouslyFocusedElements = previouslyFocusedElements.filter((entry) => entry.deref()?.isConnected);
  }
  function addPreviouslyFocusedElement(element) {
    clearDisconnectedPreviouslyFocusedElements();
    if (element && element.nodeName.toLowerCase() !== "body") {
      previouslyFocusedElements.push(new WeakRef(element));
      if (previouslyFocusedElements.length > LIST_LIMIT) previouslyFocusedElements = previouslyFocusedElements.slice(-LIST_LIMIT);
    }
  }
  function getPreviouslyFocusedElement() {
    clearDisconnectedPreviouslyFocusedElements();
    return previouslyFocusedElements[previouslyFocusedElements.length - 1]?.deref();
  }

  // A dialog without tabbable content takes the Tab stop itself. data-tabindex
  // marks the value Base UI wrote, so a user authored tabindex is left alone.
  function handleTabIndex(floatingFocusElement) {
    if (floatingFocusElement.hasAttribute("tabindex") && !floatingFocusElement.hasAttribute("data-tabindex")) return;
    if (!floatingFocusElement.getAttribute("role")?.includes("dialog")) return;
    const tabbableContent = t().focusable(floatingFocusElement).filter((element) => {
      const dataTabIndex = element.getAttribute("data-tabindex") || "";
      return t().isTabbable(element) || (element.hasAttribute("data-tabindex") && !dataTabIndex.startsWith("-"));
    });
    const tabIndex = floatingFocusElement.getAttribute("tabindex");
    if (tabbableContent.length === 0) {
      if (tabIndex !== "0") {
        floatingFocusElement.setAttribute("tabindex", "0");
        floatingFocusElement.setAttribute("data-tabindex", "0");
      }
    } else if (tabIndex !== "-1" || (floatingFocusElement.hasAttribute("data-tabindex") && floatingFocusElement.getAttribute("data-tabindex") !== "-1")) {
      floatingFocusElement.setAttribute("tabindex", "-1");
      floatingFocusElement.setAttribute("data-tabindex", "-1");
    }
  }

  // ----- the focus manager ------------------------------------------------------

  function useFloatingFocusManager(options) {
    const {
      floating,
      reference: domReference = null,
      modal = true,
      initialFocus = true,
      returnFocus = true,
      restoreFocus = false,
      closeOnFocusOut = true,
      previousFocusableElement = null,
      nextFocusableElement = null,
      getInsideElements,
      onOpenChange,
      // The element the guards wrap: the focus manager's children, the
      // floating focus element unless the popup holds a focusable list.
      guardsAround = null,
    } = options;
    // A reopened popup renders the manager again with this open's type.
    let openInteractionType = options.openInteractionType === undefined ? "" : options.openInteractionType;
    const triggers = [...(options.triggers || [domReference])].filter(Boolean);
    const doc = floating.ownerDocument;
    const floatingFocusElement = getFloatingFocusElement(floating);
    const portalNode = floating.closest("[data-base-ui-portal]");
    const ignoreInitialFocus = initialFocus === false;
    // A typeable combobox reference keeps focus in its input: no guards,
    // but the outside is still hidden.
    const isUntrappedTypeableCombobox = isTypeableCombobox(domReference) && ignoreInitialFocus;
    const self = { floating, reference: domReference, open: false };
    const cleanups = [];
    const openCleanups = [];

    let preventReturnFocus = false;
    let isPointerDown = false;
    let pointerDownOutside = false;
    let lastFocusedTabbable = null;
    let closeType = "";
    let lastInteractionType = "";
    let insideReactTree = false;
    let blurTimer = 0;
    let pointerDownTimer = 0;

    const getTabbableContent = (container = floatingFocusElement) => (container ? t().tabbable(container) : []);
    const getResolvedInsideElements = () => getInsideElements?.().filter((element) => element != null) ?? [];
    const on = (list, target, type, listener, capture) => {
      target.addEventListener(type, listener, !!capture);
      list.push(() => target.removeEventListener(type, listener, !!capture));
    };
    const isChild = (other) => other !== self && withinTree(floating, other.floating);
    const isAncestor = (other) => other !== self && withinTree(other.floating, floating);

    // Guards inside the floating tree: a modal trap, or for a non modal popup
    // in a portal the way back out through the portal's outside guards.
    const shouldRenderGuards = (modal ? !isUntrappedTypeableCombobox : true) && (portalNode != null || modal);
    let beforeGuard = null;
    let afterGuard = null;
    if (shouldRenderGuards) {
      beforeGuard = createFocusGuard("inside", (event) => {
        if (modal) {
          const els = getTabbableContent();
          enqueueFocus(els[els.length - 1]);
        } else if (portalNode) {
          preventReturnFocus = false;
          if (t().isOutsideEvent(event, portalNode)) {
            t().getNextTabbable(domReference)?.focus();
          } else {
            (previousFocusableElement || portalNode._templPortalGuards?.beforeOutside)?.focus();
          }
        }
      });
      afterGuard = createFocusGuard("inside", (event) => {
        if (modal) {
          enqueueFocus(getTabbableContent()[0]);
        } else if (portalNode) {
          if (closeOnFocusOut) preventReturnFocus = true;
          if (t().isOutsideEvent(event, portalNode)) {
            t().getPreviousTabbable(domReference)?.focus();
          } else {
            (nextFocusableElement || portalNode._templPortalGuards?.afterOutside)?.focus();
          }
        }
      });
      (guardsAround || floatingFocusElement).before(beforeGuard);
      (guardsAround || floatingFocusElement).after(afterGuard);
    }

    // Prevent Tab from escaping the modal when there is nothing tabbable.
    if (modal) {
      on(cleanups, doc, "keydown", (event) => {
        if (event.key === "Tab" && t().contains(floatingFocusElement, t().activeElement(doc)) &&
          getTabbableContent().length === 0 && !isUntrappedTypeableCombobox) {
          stopEvent(event);
        }
      });
    }

    // Close on focus out, and restore focus inside the floating tree.
    if (closeOnFocusOut) {
      const handlePointerDown = () => {
        isPointerDown = true;
        clearTimeout(pointerDownTimer);
        pointerDownTimer = setTimeout(() => {
          isPointerDown = false;
        }, 0);
      };
      const handleFocusIn = (event) => {
        const target = getTarget(event);
        if (t().isTabbable(target)) lastFocusedTabbable = target;
      };
      const handleFocusOutside = (event) => {
        const relatedTarget = event.relatedTarget;
        const currentTarget = event.currentTarget;
        const target = getTarget(event);
        // Focus lost to the body (a backdrop press): remember what had it, so a
        // confirmation dialog opened then can return focus there.
        if (modal && relatedTarget == null && target != null && t().contains(floating, target)) {
          addPreviouslyFocusedElement(target);
        }
        queueMicrotask(() => {
          const insideElements = getResolvedInsideElements();
          const portalGuards = portalNode?._templPortalGuards;
          const isRelatedFocusGuard = relatedTarget?.hasAttribute?.("data-base-ui-focus-guard") &&
            [beforeGuard, afterGuard, portalGuards?.beforeOutside, portalGuards?.afterOutside, previousFocusableElement, nextFocusableElement].includes(relatedTarget);
          const insideChild = [...instances].some((other) => isChild(other) && t().contains(other.floating, relatedTarget));
          // An ancestor's popup or its reference, not the ancestor's items.
          const toAncestor = [...instances].some((other) => isAncestor(other) &&
            ([other.floating, getFloatingFocusElement(other.floating)].includes(relatedTarget) || other.reference === relatedTarget));
          const movedToUnrelatedNode = !(
            t().contains(domReference, relatedTarget) ||
            withinTree(floating, relatedTarget) ||
            t().contains(relatedTarget, floating) ||
            t().contains(portalNode, relatedTarget) ||
            insideElements.some((element) => element === relatedTarget || t().contains(element, relatedTarget)) ||
            (relatedTarget != null && triggers.includes(relatedTarget)) ||
            triggers.some((trigger) => t().contains(trigger, relatedTarget)) ||
            isRelatedFocusGuard ||
            insideChild ||
            toAncestor
          );
          if (currentTarget === domReference && floatingFocusElement) handleTabIndex(floatingFocusElement);

          // Focus lost outside the floating tree when its element went away.
          if (restoreFocus && currentTarget !== domReference && !t().isElementVisible(target) && t().activeElement(doc) === doc.body) {
            if (isHTMLElement(floatingFocusElement)) {
              floatingFocusElement.focus();
              if (restoreFocus === "popup") {
                requestAnimationFrame(() => floatingFocusElement.focus());
                return;
              }
            }
            const tabbableContent = getTabbableContent();
            const prevTabbable = lastFocusedTabbable;
            const nodeToFocus = (prevTabbable && tabbableContent.includes(prevTabbable) ? prevTabbable : null) ||
              tabbableContent[tabbableContent.length - 1] || floatingFocusElement;
            if (isHTMLElement(nodeToFocus)) nodeToFocus.focus();
          }

          if (insideReactTree) {
            insideReactTree = false;
            return;
          }

          // Focus moved out of the floating tree with no portal guard to handle it.
          if ((isUntrappedTypeableCombobox ? true : !modal) && relatedTarget && movedToUnrelatedNode && !isPointerDown &&
            (isUntrappedTypeableCombobox || relatedTarget !== getPreviouslyFocusedElement())) {
            preventReturnFocus = true;
            onOpenChange?.(false, "focus-out", event);
          }
        });
      };
      // Focus leaving through a portaled child still counts as inside.
      const markInsideReactTree = () => {
        if (pointerDownOutside) return;
        insideReactTree = true;
        clearTimeout(blurTimer);
        blurTimer = setTimeout(() => {
          insideReactTree = false;
        }, 0);
      };
      if (isHTMLElement(domReference)) {
        on(cleanups, domReference, "focusout", handleFocusOutside);
        on(cleanups, domReference, "pointerdown", handlePointerDown);
      }
      on(cleanups, floating, "focusin", handleFocusIn);
      on(cleanups, floating, "focusout", handleFocusOutside);
      if (portalNode) on(cleanups, floating, "focusout", markInsideReactTree, true);
    }

    // The effects that run while open. A popup that opens again during its
    // exit animation runs them again on the same manager.
    function open(nextOpenInteractionType) {
      if (nextOpenInteractionType !== undefined) openInteractionType = nextOpenInteractionType;
      self.open = true;
      // The portal's outside guards render first, they count as inside below.
      window.templ.portal.setFocusManagerState(portalNode, { ...portalNode?._templFocusState, open: true });
      // Track pointer and keyboard interactions while open.
      on(openCleanups, doc, "pointerdown", (event) => {
        const target = getTarget(event);
        const insideElements = getResolvedInsideElements();
        const pointerTargetInside = t().contains(floating, target) || t().contains(domReference, target) ||
          t().contains(portalNode, target) || insideElements.some((element) => element === target || t().contains(element, target));
        pointerDownOutside = !pointerTargetInside;
        lastInteractionType = event.pointerType || "keyboard";
        if (target?.closest?.(`[${CLICK_TRIGGER_IDENTIFIER}]`)) {
          isPointerDown = true;
          clearTimeout(pointerDownTimer);
          pointerDownTimer = setTimeout(() => {
            isPointerDown = false;
          }, 0);
        }
      }, true);
      const clearPointerDownOutside = () => {
        pointerDownOutside = false;
      };
      on(openCleanups, doc, "pointerup", clearPointerDownOutside, true);
      on(openCleanups, doc, "pointercancel", clearPointerDownOutside, true);
      on(openCleanups, doc, "keydown", () => {
        lastInteractionType = "keyboard";
      }, true);
      openCleanups.push(clearPointerDownOutside);

      // Hide everything outside the floating tree from assistive tech while open.
      const nestedPortalNodes = Array.from(portalNode?.querySelectorAll("[data-base-ui-portal]") || []);
      const portalGuards = portalNode?._templPortalGuards;
      const insideElements = [
        floating, ...nestedPortalNodes, beforeGuard, afterGuard, portalGuards?.beforeOutside, portalGuards?.afterOutside,
        ...getResolvedInsideElements(), previousFocusableElement, nextFocusableElement,
        isUntrappedTypeableCombobox ? domReference : null,
      ].filter((x) => x != null);
      const ariaHiddenCleanup = window.templ.markOthers(insideElements, { ariaHidden: modal || isUntrappedTypeableCombobox, mark: false });
      const markerCleanup = window.templ.markOthers([floating, ...nestedPortalNodes].filter((x) => x != null));
      openCleanups.push(() => {
        markerCleanup();
        ariaHiddenCleanup();
      });

      // Focus the initial element.
      const previouslyFocusedElement = t().activeElement(doc);
      queueMicrotask(() => {
        const resolvedInitialFocus = typeof initialFocus === "function" ? initialFocus(openInteractionType || "") : initialFocus;
        if (resolvedInitialFocus === undefined || resolvedInitialFocus === false) return;
        if (t().contains(floatingFocusElement, previouslyFocusedElement)) return;
        const getDefaultFocusElement = () => getTabbableContent(floatingFocusElement)[0] || floatingFocusElement;
        let elToFocus = resolvedInitialFocus === true || resolvedInitialFocus === null ? getDefaultFocusElement() : resolvedInitialFocus;
        elToFocus = elToFocus || getDefaultFocusElement();
        const hadFocusInside = t().contains(floatingFocusElement, t().activeElement(doc));
        enqueueFocus(elToFocus, {
          preventScroll: elToFocus === floatingFocusElement,
          shouldFocus() {
            if (!self.open) return false;
            if (hadFocusInside) return true;
            const currentActiveElement = t().activeElement(doc);
            return !(currentActiveElement !== elToFocus && t().contains(floatingFocusElement, currentActiveElement));
          },
        });
      });
    }

    // Return focus targets, restored on unmount. Only a null interaction
    // type is a programmatic open.
    const elementFocusedBeforeOpen = t().activeElement(doc);
    const preferPreviousFocus = openInteractionType == null;
    addPreviouslyFocusedElement(elementFocusedBeforeOpen);

    function getReturnElement() {
      let resolved = typeof returnFocus === "function" ? returnFocus(closeType) : returnFocus;
      if (resolved === undefined || resolved === false) return null;
      if (resolved === null) resolved = true;
      const referenceReturnElement = domReference?.isConnected ? domReference : null;
      const previousReturnElement = elementFocusedBeforeOpen?.isConnected && elementFocusedBeforeOpen.nodeName.toLowerCase() !== "body"
        ? elementFocusedBeforeOpen
        : null;
      let defaultReturnElement = preferPreviousFocus
        ? previousReturnElement || referenceReturnElement
        : referenceReturnElement || previousReturnElement;
      if (!defaultReturnElement) defaultReturnElement = getPreviouslyFocusedElement() || null;
      if (typeof resolved === "boolean") return defaultReturnElement;
      return resolved || defaultReturnElement || null;
    }

    // The portal renders its outside guards for a non modal manager.
    window.templ.portal.setFocusManagerState(portalNode, {
      modal,
      closeOnFocusOut,
      open: false,
      onOpenChange,
      domReference,
      beforeInside: beforeGuard,
      afterInside: afterGuard,
    });
    if (floatingFocusElement) handleTabIndex(floatingFocusElement);
    instances.add(self);
    open();

    function close(details = {}) {
      if (!self.open) return;
      self.open = false;
      closeType = getEventType(details.event, lastInteractionType);
      if (details.reason === "trigger-hover" && details.event?.type === "mouseleave") preventReturnFocus = true;
      if (details.reason === "outside-press") {
        // An outside press returns focus without scrolling where the browser
        // supports preventScroll, everywhere current.
        const event = details.event;
        if (details.nested || (event && (isVirtualClick(event) || isVirtualPointerEvent(event)))) {
          preventReturnFocus = false;
        } else {
          let isPreventScrollSupported = false;
          doc.createElement("div").focus({
            get preventScroll() {
              isPreventScrollSupported = true;
              return false;
            },
          });
          preventReturnFocus = !isPreventScrollSupported;
        }
      }
      openCleanups.splice(0).forEach((cleanup) => cleanup());
      window.templ.portal.setFocusManagerState(portalNode, { ...portalNode?._templFocusState, open: false });
      // Safari may scroll to the bottom when an input inside the popup keeps
      // focus while it unmounts.
      const activeEl = t().activeElement(doc);
      if (webkit && isHTMLElement(activeEl) && isTypeableElement(activeEl) && t().contains(floating, activeEl)) activeEl.blur();
    }

    function unmount() {
      if (self.open) close();
      instances.delete(self);
      cleanups.splice(0).forEach((cleanup) => cleanup());
      clearTimeout(blurTimer);
      clearTimeout(pointerDownTimer);
      window.templ.portal.setFocusManagerState(portalNode, null);
      beforeGuard?.remove();
      afterGuard?.remove();

      const activeEl = t().activeElement(doc);
      const isFocusInsideFloatingTree = t().contains(floating, activeEl) ||
        getResolvedInsideElements().some((element) => element === activeEl || t().contains(element, activeEl)) ||
        [...instances].some((other) => isChild(other) && t().contains(other.floating, activeEl));
      const returnElement = getReturnElement();
      queueMicrotask(() => {
        const tabbableReturnElement = getFirstTabbableElement(returnElement);
        const hasExplicitReturnFocus = typeof returnFocus !== "boolean";
        // Focus that moved elsewhere after mount is respected.
        if (returnFocus && !preventReturnFocus && isHTMLElement(tabbableReturnElement) &&
          (!hasExplicitReturnFocus && tabbableReturnElement !== activeEl && activeEl !== doc.body ? isFocusInsideFloatingTree : true)) {
          tabbableReturnElement.focus({ preventScroll: true });
        }
        preventReturnFocus = false;
        clearDisconnectedPreviouslyFocusedElements();
      });
    }

    // beforeGuard is the source's beforeContentFocusGuardRef.
    return { open, close, unmount, beforeGuard };
  }

  window.templ = window.templ || {};
  window.templ.focusManager = { useFloatingFocusManager, createFocusGuard, enqueueFocus };
})();

// components/baseui/lifecycle.js
// Pendant of React mount and unmount. A Base UI part mounts when React renders
// it and unmounts when React removes it, and a portaled part stays mounted as
// long as the element that rendered it. Here one MutationObserver does that
// for every component: init(el) runs once when an element matching the
// selector appears, destroy(el) once when it is gone.
(function () {
  "use strict";

  const components = [];

  // A portaled subtree lives as long as its declaration site (portal owner),
  // like React unmounts a portal with the component that rendered it.
  function mounted(el) {
    if (!el.isConnected) return false;
    for (let node = el; node; node = node.parentElement) {
      if (node._templPortalOwner && !mounted(node._templPortalOwner)) return false;
    }
    return true;
  }

  function sync(component) {
    for (const el of component.elements) {
      if (mounted(el)) continue;
      component.elements.delete(el);
      component.destroy?.(el);
    }
    for (const el of document.querySelectorAll(component.selector)) {
      if (component.elements.has(el) || !mounted(el)) continue;
      component.elements.add(el);
      component.init?.(el);
    }
  }

  // register(selector, { init, destroy }) wires every matching element that
  // is in the document now and every one that appears later.
  function register(selector, { init, destroy } = {}) {
    const component = { selector, init, destroy, elements: new Set() };
    components.push(component);
    if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", () => sync(component));
    else sync(component);
  }

  new MutationObserver(() => components.forEach(sync))
    .observe(document.documentElement, { childList: true, subtree: true });

  window.templ = window.templ || {};
  window.templ.lifecycle = { register };
})();

// components/baseui/mark_others.js
// Port of @base-ui/react floating-ui-react/utils/markOthers.ts (1.6.0), itself a
// fork of aria-hidden with a conditional aria-hidden. markOthers(avoid, options)
// marks every element outside the avoided subtrees and returns the undo.
//
//   ariaHidden  sets aria-hidden="true" outside (modal popups)
//   inert       sets inert outside
//   mark        sets the data-base-ui-inert marker outside, which useDismiss
//               reads to tell elements injected after opening (default true)
//
// Counted per element, so nested popups and several open ones undo cleanly,
// and elements that were hidden before are left as they were.
(function () {
  "use strict";

  const markerName = "data-base-ui-inert";
  let counters = { inert: new WeakMap(), "aria-hidden": new WeakMap() };
  let uncontrolledElementsSets = { inert: new WeakSet(), "aria-hidden": new WeakSet() };
  let markerCounterMap = new WeakMap();
  let lockCount = 0;

  const isShadowRoot = (node) => typeof ShadowRoot !== "undefined" && node instanceof ShadowRoot;

  function unwrapHost(node) {
    if (!node) return null;
    return isShadowRoot(node) ? node.host : unwrapHost(node.parentNode);
  }

  function correctElements(parent, targets) {
    return targets.map((target) => {
      if (parent.contains(target)) return target;
      const correctedTarget = unwrapHost(target);
      if (parent.contains(correctedTarget)) return correctedTarget;
      return null;
    }).filter((x) => x != null);
  }

  function buildKeepSet(targets) {
    const keep = new Set();
    targets.forEach((target) => {
      let node = target;
      while (node && !keep.has(node)) {
        keep.add(node);
        node = node.parentNode;
      }
    });
    return keep;
  }

  function collectOutsideElements(root, keepElements, stopElements) {
    const outside = [];
    const walk = (parent) => {
      if (!parent || stopElements.has(parent)) return;
      Array.from(parent.children).forEach((node) => {
        if (node.nodeName.toLowerCase() === "script") return;
        if (keepElements.has(node)) walk(node);
        else outside.push(node);
      });
    };
    walk(root);
    return outside;
  }

  function applyAttributeToOthers(uncorrectedAvoidElements, body, ariaHidden, inert, { mark = true }) {
    let controlAttribute = null;
    if (inert) controlAttribute = "inert";
    else if (ariaHidden) controlAttribute = "aria-hidden";
    let counterMap = null;
    let uncontrolledElementsSet = null;
    const avoidElements = correctElements(body, uncorrectedAvoidElements);
    const markerTargets = mark ? collectOutsideElements(body, buildKeepSet(avoidElements), new Set(avoidElements)) : [];
    const hiddenElements = [];
    const markedElements = [];
    if (controlAttribute) {
      const map = counters[controlAttribute];
      uncontrolledElementsSet = uncontrolledElementsSets[controlAttribute];
      counterMap = map;
      const ariaLiveElements = correctElements(body, Array.from(body.querySelectorAll("[aria-live]")));
      const controlElements = avoidElements.concat(ariaLiveElements);
      const controlTargets = collectOutsideElements(body, buildKeepSet(controlElements), new Set(controlElements));
      controlTargets.forEach((node) => {
        const attr = node.getAttribute(controlAttribute);
        const alreadyHidden = attr !== null && attr !== "false";
        const counterValue = (map.get(node) || 0) + 1;
        map.set(node, counterValue);
        hiddenElements.push(node);
        if (counterValue === 1 && alreadyHidden) uncontrolledElementsSet.add(node);
        if (!alreadyHidden) node.setAttribute(controlAttribute, controlAttribute === "inert" ? "" : "true");
      });
    }
    if (mark) {
      markerTargets.forEach((node) => {
        const markerValue = (markerCounterMap.get(node) || 0) + 1;
        markerCounterMap.set(node, markerValue);
        markedElements.push(node);
        if (markerValue === 1) node.setAttribute(markerName, "");
      });
    }
    lockCount += 1;
    return () => {
      if (counterMap) {
        hiddenElements.forEach((element) => {
          const counterValue = (counterMap.get(element) || 0) - 1;
          counterMap.set(element, counterValue);
          if (!counterValue) {
            if (!uncontrolledElementsSet?.has(element) && controlAttribute) element.removeAttribute(controlAttribute);
            uncontrolledElementsSet?.delete(element);
          }
        });
      }
      if (mark) {
        markedElements.forEach((element) => {
          const markerValue = (markerCounterMap.get(element) || 0) - 1;
          markerCounterMap.set(element, markerValue);
          if (!markerValue) element.removeAttribute(markerName);
        });
      }
      lockCount -= 1;
      if (!lockCount) {
        counters = { inert: new WeakMap(), "aria-hidden": new WeakMap() };
        uncontrolledElementsSets = { inert: new WeakSet(), "aria-hidden": new WeakSet() };
        markerCounterMap = new WeakMap();
      }
    };
  }

  function markOthers(avoidElements, options = {}) {
    const { ariaHidden = false, inert = false, mark = true } = options;
    const body = avoidElements[0].ownerDocument.body;
    return applyAttributeToOthers(avoidElements, body, ariaHidden, inert, { mark });
  }

  window.templ = window.templ || {};
  window.templ.markOthers = markOthers;
})();

// components/baseui/portal.js
// Port of @base-ui/react floating-ui-react/components/FloatingPortal.tsx (1.6.0).
//
// render(node) moves a [data-base-ui-portal] node to its container while its
// component stays where it was declared. That declaration site is the portal
// owner, and lifecycle.js unmounts the portaled subtree once the owner leaves
// the document. The container is the portal node of an enclosing portal, or
// <body> (useFloatingPortalNode). remove(node) unmounts it with its guards.
//
// setFocusManagerState(node, state) is the source's portal context: a non
// modal focus manager inside the node gets the outside guards at the
// declaration site, which move Tab between the trigger and the portaled
// content, and the node's tabbables are taken out of the tab order while
// focus is outside of it.
(function () {
  "use strict";

  let uid = 0;
  const t = () => window.templ.tabbable;

  // The React tree pendant: the parent of a portal node is where it was
  // declared, of every other node its DOM parent.
  function treeParent(node) {
    return node._templPortalOwner || node.parentNode;
  }

  function containerFor(node) {
    return node._templPortalOwner?.closest("[data-base-ui-portal]") || document.body;
  }

  // Appends on every open, so paint order follows open order like the
  // portal nodes React creates on mount. A portal declared right inside its
  // container's node (a Portal part around a submenu) comes before its own
  // outside guards there, as upstream renders it. target is createPortal's
  // container when it is not the portal node's own (NavigationMenu's
  // viewport).
  function render(node, target) {
    if (!node._templPortalOwner) node._templPortalOwner = node.parentElement;
    if (target) {
      target.appendChild(node);
      return;
    }
    if (!node.id) node.id = "templ-portal-" + ++uid;
    const container = containerFor(node);
    const owner = node._templPortalOwner;
    if (container !== document.body && owner.parentElement === container) container.insertBefore(node, node._templPortalGuards?.beforeOutside || owner);
    else container.appendChild(node);
  }

  function remove(node) {
    if (!node) return;
    setFocusManagerState(node, null);
    if (node.isConnected) node.remove();
  }

  // Tabbables inside the node count only once focus entered it, through an
  // outside guard or a pointer.
  function onFocus(event) {
    const node = event.currentTarget;
    if (!event.relatedTarget || !t().isOutsideEvent(event)) return;
    if (event.type === "focusin") {
      if (node._templFocusInsideDisabled) {
        t().enableFocusInside(node);
        node._templFocusInsideDisabled = false;
      }
    } else {
      t().disableFocusInside(node);
      node._templFocusInsideDisabled = true;
    }
  }

  function createOutsideGuards(node) {
    const { createFocusGuard } = window.templ.focusManager;
    const state = () => node._templFocusState;
    const beforeOutside = createFocusGuard("outside", (event) => {
      if (t().isOutsideEvent(event, node)) {
        state()?.beforeInside?.focus();
      } else {
        const domReference = state()?.domReference;
        if (domReference) t().getPreviousTabbable(domReference)?.focus();
      }
    });
    const afterOutside = createFocusGuard("outside", (event) => {
      if (t().isOutsideEvent(event, node)) {
        state()?.afterInside?.focus();
      } else {
        const domReference = state()?.domReference;
        if (domReference) t().getNextTabbable(domReference)?.focus();
        if (state()?.closeOnFocusOut) state()?.onOpenChange?.(false, "focus-out", event);
      }
    });
    const owns = document.createElement("span");
    owns.setAttribute("aria-owns", node.id);
    owns.style.cssText = "clip-path:inset(50%);position:fixed;top:0;left:0";
    node._templPortalOwner?.before(beforeOutside, owns, afterOutside);
    return { beforeOutside, afterOutside, owns };
  }

  function setFocusManagerState(node, state) {
    if (!node) return;
    node._templFocusState = state;
    const nonModal = !!state && !state.modal;

    if (nonModal && !node._templFocusListening) {
      node.addEventListener("focusin", onFocus, true);
      node.addEventListener("focusout", onFocus, true);
      node._templFocusListening = true;
    } else if (!nonModal && node._templFocusListening) {
      node.removeEventListener("focusin", onFocus, true);
      node.removeEventListener("focusout", onFocus, true);
      node._templFocusListening = false;
    }
    // Tabbable again before the focus manager's queued initial focus runs.
    if (state?.open && node._templFocusInsideDisabled) {
      t().enableFocusInside(node);
      node._templFocusInsideDisabled = false;
    }

    const shouldRenderGuards = nonModal && state.open;
    if (shouldRenderGuards && !node._templPortalGuards) {
      node._templPortalGuards = createOutsideGuards(node);
    } else if (!shouldRenderGuards && node._templPortalGuards) {
      const { beforeOutside, afterOutside, owns } = node._templPortalGuards;
      beforeOutside.remove();
      owns.remove();
      afterOutside.remove();
      node._templPortalGuards = null;
    }
  }

  window.templ = window.templ || {};
  window.templ.portal = { render, remove, setFocusManagerState, treeParent };
})();

// components/baseui/scroll_lock.js
// Port of @base-ui/utils/useScrollLock.ts. Helpers from useTimeout.ts,
// useAnimationFrame.ts, platform/os.ts, platform/engine.ts, @floating-ui/utils/dom,
// and @base-ui/react/utils/useAnchoredPopupScrollLock.ts live here as well.
(function () {
  "use strict";

  // @base-ui/utils/platform/{os,engine}.ts
  const lowerPlatform = navigator.platform.toLowerCase();
  const ios = /^i(os$|p)/.test(lowerPlatform) ||
    (lowerPlatform === "macintel" && navigator.maxTouchPoints > 1);
  const webkit = typeof CSS !== "undefined" && !!CSS.supports?.("-webkit-backdrop-filter:none");

  const ownerDocument = (referenceElement) => referenceElement?.ownerDocument || document;
  const ownerWindow = (referenceElement) =>
    (referenceElement?.nodeType === 9 ? referenceElement : ownerDocument(referenceElement)).defaultView || window;

  // @floating-ui/utils/dom: isOverflowElement
  function isOverflowElement(element) {
    const { overflow, overflowX, overflowY, display } = ownerWindow(element).getComputedStyle(element);
    return /auto|scroll|overlay|hidden|clip/.test(overflow + overflowY + overflowX) &&
      display !== "inline" && display !== "contents";
  }

  // @base-ui/utils/useTimeout.ts (the imperative helper; no React lifecycle).
  class Timeout {
    static create() { return new Timeout(); }
    currentId = 0;
    start(delay, fn) {
      this.clear();
      this.currentId = setTimeout(() => {
        this.currentId = 0;
        fn();
      }, delay);
    }
    isStarted() { return this.currentId !== 0; }
    clear = () => {
      if (this.currentId !== 0) {
        clearTimeout(this.currentId);
        this.currentId = 0;
      }
    };
  }

  // @base-ui/utils/useAnimationFrame.ts, including its production scheduler.
  class Scheduler {
    callbacks = [];
    callbacksCount = 0;
    nextId = 1;
    startId = 1;
    isScheduled = false;
    tick = (timestamp) => {
      this.isScheduled = false;
      const currentCallbacks = this.callbacks;
      const currentCallbacksCount = this.callbacksCount;
      this.callbacks = [];
      this.callbacksCount = 0;
      this.startId = this.nextId;
      if (currentCallbacksCount > 0) {
        for (let i = 0; i < currentCallbacks.length; i += 1) {
          currentCallbacks[i]?.(timestamp);
        }
      }
    };
    request(fn) {
      const id = this.nextId;
      this.nextId += 1;
      this.callbacks.push(fn);
      this.callbacksCount += 1;
      if (!this.isScheduled) {
        requestAnimationFrame(this.tick);
        this.isScheduled = true;
      }
      return id;
    }
    cancel(id) {
      const index = id - this.startId;
      if (index < 0 || index >= this.callbacks.length || this.callbacks[index] === null) return;
      this.callbacks[index] = null;
      this.callbacksCount -= 1;
    }
  }
  const scheduler = new Scheduler();
  class AnimationFrame {
    static create() { return new AnimationFrame(); }
    static request(fn) { return scheduler.request(fn); }
    static cancel(id) { scheduler.cancel(id); }
    currentId = null;
    request(fn) {
      this.cancel();
      this.currentId = scheduler.request(() => {
        this.currentId = null;
        fn();
      });
    }
    cancel = () => {
      if (this.currentId !== null) {
        scheduler.cancel(this.currentId);
        this.currentId = null;
      }
    };
  }

  let originalHtmlStyles = {};
  let originalBodyStyles = {};
  let originalHtmlScrollBehavior = '';

  // The viewport's overflow comes from <html> when it establishes its own scroll container, and
  // propagates from <body> otherwise. An `overflow` style on the other element doesn't lock the page.
  function getViewportScroller(html, body) {
    return isOverflowElement(html) ? html : body;
  }

  function isPageScrollLocked(win, html, body) {
    return /hidden|clip/.test(win.getComputedStyle(getViewportScroller(html, body)).overflowY);
  }

  function hasInsetScrollbars(referenceElement) {
    if (typeof document === 'undefined') {
      return false;
    }
    const doc = ownerDocument(referenceElement);
    const win = ownerWindow(doc);
    return win.innerWidth - doc.documentElement.clientWidth > 0;
  }

  function supportsStableScrollbarGutter(referenceElement) {
    const supported =
      typeof CSS !== 'undefined' && CSS.supports && CSS.supports('scrollbar-gutter', 'stable');

    if (!supported || typeof document === 'undefined') {
      return false;
    }

    const doc = ownerDocument(referenceElement);
    const html = doc.documentElement;
    const body = doc.body;

    const scrollContainer = getViewportScroller(html, body);

    const originalScrollContainerOverflowY = scrollContainer.style.overflowY;
    const originalHtmlStyleGutter = html.style.scrollbarGutter;

    html.style.scrollbarGutter = 'stable';

    scrollContainer.style.overflowY = 'scroll';
    const before = scrollContainer.offsetWidth;

    scrollContainer.style.overflowY = 'hidden';
    const after = scrollContainer.offsetWidth;

    scrollContainer.style.overflowY = originalScrollContainerOverflowY;
    html.style.scrollbarGutter = originalHtmlStyleGutter;

    return before === after;
  }

  function preventScrollOverlayScrollbars(referenceElement) {
    const doc = ownerDocument(referenceElement);
    const html = doc.documentElement;
    const body = doc.body;

    // If an `overflow` style is present on <html>, we need to lock it, because a lock on <body>
    // won't have any effect.
    // But if <body> has an `overflow` style (like `overflow-x: hidden`), we need to lock it
    // instead, as sticky elements shift otherwise.
    const elementToLock = getViewportScroller(html, body);
    const originalElementToLockStyles = {
      overflowY: elementToLock.style.overflowY,
      overflowX: elementToLock.style.overflowX,
    };

    Object.assign(elementToLock.style, {
      overflowY: 'hidden',
      overflowX: 'hidden',
    });

    return () => {
      Object.assign(elementToLock.style, originalElementToLockStyles);
    };
  }

  function preventScrollInsetScrollbars(referenceElement) {
    const doc = ownerDocument(referenceElement);
    const html = doc.documentElement;
    const body = doc.body;
    const win = ownerWindow(html);

    let scrollTop = 0;
    let scrollLeft = 0;
    let updateGutterOnly = false;
    const resizeFrame = AnimationFrame.create();

    // Pinch-zoom in Safari causes a shift. Just don't lock scroll if there's any pinch-zoom.
    if (webkit && (win.visualViewport?.scale ?? 1) !== 1) {
      return () => {};
    }

    function lockScroll() {
      /* DOM reads: */

      const htmlStyles = win.getComputedStyle(html);
      const bodyStyles = win.getComputedStyle(body);
      const htmlScrollbarGutterValue = htmlStyles.scrollbarGutter || '';
      const hasBothEdges = htmlScrollbarGutterValue.includes('both-edges');
      const scrollbarGutterValue = hasBothEdges ? 'stable both-edges' : 'stable';

      scrollTop = html.scrollTop;
      scrollLeft = html.scrollLeft;

      originalHtmlStyles = {
        scrollbarGutter: html.style.scrollbarGutter,
        overflowY: html.style.overflowY,
        overflowX: html.style.overflowX,
      };
      originalHtmlScrollBehavior = html.style.scrollBehavior;

      originalBodyStyles = {
        position: body.style.position,
        height: body.style.height,
        width: body.style.width,
        boxSizing: body.style.boxSizing,
        overflowY: body.style.overflowY,
        overflowX: body.style.overflowX,
        scrollBehavior: body.style.scrollBehavior,
      };

      const isScrollableY = html.scrollHeight > html.clientHeight;
      const isScrollableX = html.scrollWidth > html.clientWidth;
      const hasConstantOverflowY =
        htmlStyles.overflowY === 'scroll' || bodyStyles.overflowY === 'scroll';
      const hasConstantOverflowX =
        htmlStyles.overflowX === 'scroll' || bodyStyles.overflowX === 'scroll';

      // Values can be negative in Firefox
      const scrollbarWidth = Math.max(0, win.innerWidth - body.clientWidth);
      const scrollbarHeight = Math.max(0, win.innerHeight - body.clientHeight);

      // Avoid shift due to the default <body> margin. This does cause elements to be clipped
      // with whitespace. Warn if <body> has margins?
      const marginY = parseFloat(bodyStyles.marginTop) + parseFloat(bodyStyles.marginBottom);
      const marginX = parseFloat(bodyStyles.marginLeft) + parseFloat(bodyStyles.marginRight);
      const elementToLock = getViewportScroller(html, body);

      updateGutterOnly = supportsStableScrollbarGutter(referenceElement);

      /*
       * DOM writes:
       * Do not read the DOM past this point!
       */

      if (updateGutterOnly) {
        html.style.scrollbarGutter = scrollbarGutterValue;
        elementToLock.style.overflowY = 'hidden';
        elementToLock.style.overflowX = 'hidden';
        return;
      }

      Object.assign(html.style, {
        scrollbarGutter: scrollbarGutterValue,
        overflowY: 'hidden',
        overflowX: 'hidden',
      });

      if (isScrollableY || hasConstantOverflowY) {
        html.style.overflowY = 'scroll';
      }
      if (isScrollableX || hasConstantOverflowX) {
        html.style.overflowX = 'scroll';
      }

      Object.assign(body.style, {
        position: 'relative',
        height:
          marginY || scrollbarHeight ? `calc(100dvh - ${marginY + scrollbarHeight}px)` : '100dvh',
        width: marginX || scrollbarWidth ? `calc(100vw - ${marginX + scrollbarWidth}px)` : '100vw',
        boxSizing: 'border-box',
        // Assign the longhands that `cleanup` restores, so nothing is left behind.
        overflowY: 'hidden',
        overflowX: 'hidden',
        scrollBehavior: 'unset',
      });

      body.scrollTop = scrollTop;
      body.scrollLeft = scrollLeft;
      html.setAttribute('data-base-ui-scroll-locked', '');
      html.style.scrollBehavior = 'unset';
    }

    function cleanup() {
      Object.assign(html.style, originalHtmlStyles);
      Object.assign(body.style, originalBodyStyles);

      if (!updateGutterOnly) {
        html.scrollTop = scrollTop;
        html.scrollLeft = scrollLeft;
        html.removeAttribute('data-base-ui-scroll-locked');
        html.style.scrollBehavior = originalHtmlScrollBehavior;
      }
    }

    function handleResize() {
      cleanup();
      resizeFrame.request(lockScroll);
    }

    lockScroll();
    win.addEventListener('resize', handleResize);

    return () => {
      resizeFrame.cancel();
      cleanup();
      // Sometimes this cleanup can run after test teardown because it is called
      // in a `setTimeout(fn, 0)`. Guard the returned cleanup to avoid calling
      // `removeEventListener` when it is no longer available in tests.
      if (typeof win.removeEventListener === 'function') {
        win.removeEventListener('resize', handleResize);
      }
    };
  }

  class ScrollLocker {
    lockCount = 0;
    restore = null;
    timeoutLock = Timeout.create();
    timeoutUnlock = Timeout.create();

    acquire(referenceElement) {
      this.lockCount += 1;
      if (this.lockCount === 1 && this.restore === null) {
        this.timeoutLock.start(0, () => this.lock(referenceElement));
      }
      return this.release;
    }

    release = () => {
      this.lockCount -= 1;
      if (this.lockCount === 0 && this.restore) {
        this.timeoutUnlock.start(0, this.unlock);
      }
    };

    unlock = () => {
      if (this.lockCount === 0 && this.restore) {
        this.restore?.();
        this.restore = null;
      }
    };

    lock(referenceElement) {
      if (this.lockCount === 0 || this.restore !== null) {
        return;
      }

      const doc = ownerDocument(referenceElement);
      const html = doc.documentElement;
      const body = doc.body;
      const win = ownerWindow(html);

      // The page is already locked, either by the site author or by a non-Base UI overlay that
      // hasn't cleaned up yet. Leave it alone and wait for the lock to clear before taking over,
      // otherwise we'd snapshot the locked state and restore it after our own lock is released.
      if (isPageScrollLocked(win, html, body)) {
        const observer = new win.MutationObserver(() => {
          if (isPageScrollLocked(win, html, body)) {
            return;
          }
          observer.disconnect();
          this.restore = null;
          this.lock(referenceElement);
        });

        // Watch every attribute: locks are applied through inline styles, classes, or attributes
        // paired with a stylesheet (`data-scroll-locked` in react-remove-scroll, for example).
        const options = { attributes: true };

        observer.observe(html, options);
        observer.observe(body, options);

        this.restore = () => observer.disconnect();
        return;
      }

      const hasOverlayScrollbars = ios || !hasInsetScrollbars(referenceElement);

      // On iOS, scroll locking does not work if the navbar is collapsed. Due to numerous
      // side effects and bugs that arise on iOS, it must be researched extensively before
      // being enabled to ensure it doesn't cause the following issues:
      // - Textboxes must scroll into view when focused, nor cause a glitchy scroll animation.
      // - The navbar must not force itself into view and cause layout shift.
      // - Scroll containers must not flicker upon closing a popup when it has an exit animation.
      this.restore = hasOverlayScrollbars
        ? preventScrollOverlayScrollbars(referenceElement)
        : preventScrollInsetScrollbars(referenceElement);
    }
  }

  const SCROLL_LOCKER = new ScrollLocker();

  // @base-ui/react/utils/useAnchoredPopupScrollLock.ts: run after positioning.
  const VIEWPORT_WIDTH_TOLERANCE_PX = 20;
  function anchoredPopupScrollLock(enabled, touchOpen, positionerElement, referenceElement) {
    let touchOpenShouldLockScroll = false;
    if (enabled && touchOpen && positionerElement != null) {
      const viewportWidth = ownerDocument(positionerElement).documentElement.clientWidth;
      const popupWidth = positionerElement.offsetWidth;
      touchOpenShouldLockScroll = viewportWidth > 0 && popupWidth > 0 &&
        popupWidth >= viewportWidth - VIEWPORT_WIDTH_TOLERANCE_PX;
    }
    return enabled && (!touchOpen || touchOpenShouldLockScroll)
      ? SCROLL_LOCKER.acquire(referenceElement)
      : () => {};
  }

  window.templ = window.templ || {};
  window.templ.scrollLock = {
    acquire: (referenceElement) => SCROLL_LOCKER.acquire(referenceElement),
    anchoredPopup: anchoredPopupScrollLock,
  };
})();

// components/baseui/tabbable.js
// Port of @base-ui/react floating-ui-react/utils/tabbable.ts (1.6.0), Base UI's
// own tabbable implementation, with activeElement and contains from
// utils/element.ts and isElementVisible from utils/composite.ts.
(function () {
  "use strict";

  const CANDIDATE_SELECTOR = 'a[href],button,input,select,textarea,summary,details,iframe,object,embed,[tabindex],[contenteditable]:not([contenteditable="false"]),audio[controls],video[controls]';

  const nodeName = (element) => (element?.nodeName || "").toLowerCase();
  const isShadowRoot = (node) => typeof ShadowRoot !== "undefined" && node instanceof ShadowRoot;
  const isHTMLElement = (node) => node instanceof HTMLElement;

  // The focused element, through open shadow roots.
  function activeElement(doc) {
    let element = doc.activeElement;
    while (element?.shadowRoot?.activeElement != null) element = element.shadowRoot.activeElement;
    return element;
  }

  function contains(parent, child) {
    if (!parent || !child) return false;
    const rootNode = child.getRootNode?.();
    if (parent.contains(child)) return true;
    if (rootNode && isShadowRoot(rootNode)) {
      let next = child;
      while (next) {
        if (parent === next) return true;
        next = next.parentNode || next.host;
      }
    }
    return false;
  }

  function isHiddenByStyles(styles) {
    return styles.visibility === "hidden" || styles.visibility === "collapse";
  }

  function isElementVisible(element, styles = element ? getComputedStyle(element) : null) {
    if (!element || !element.isConnected || !styles || isHiddenByStyles(styles)) return false;
    if (typeof element.checkVisibility === "function") return element.checkVisibility();
    return styles.display !== "none" && styles.display !== "contents";
  }

  function getParentElement(element) {
    if (element.assignedSlot) return element.assignedSlot;
    if (element.parentElement) return element.parentElement;
    const rootNode = element.getRootNode();
    return isShadowRoot(rootNode) ? rootNode.host : null;
  }

  function getDetailsSummary(details) {
    for (const child of Array.from(details.children)) {
      if (nodeName(child) === "summary") return child;
    }
    return null;
  }

  function isWithinOpenDetailsSummary(element, details) {
    const summary = getDetailsSummary(details);
    return !!summary && (element === summary || contains(summary, element));
  }

  function isFocusableCandidate(element) {
    const name = element ? nodeName(element) : "";
    return element != null && element.matches(CANDIDATE_SELECTOR) &&
      (name !== "summary" || (element.parentElement != null && nodeName(element.parentElement) === "details" && getDetailsSummary(element.parentElement) === element)) &&
      (name !== "details" || getDetailsSummary(element) == null) &&
      (name !== "input" || element.type !== "hidden");
  }

  function isVisibleInTabbableTree(element, isAncestor) {
    const styles = getComputedStyle(element);
    if (!isAncestor) return isElementVisible(element, styles);
    return styles.display !== "none";
  }

  function isFocusableElement(element) {
    if (!isFocusableCandidate(element) || !element.isConnected || element.matches(":disabled")) return false;
    for (let current = element; current; current = getParentElement(current)) {
      const isAncestor = current !== element;
      const isSlot = nodeName(current) === "slot";
      if (current.hasAttribute("inert")) return false;
      if ((isAncestor && nodeName(current) === "details" && !current.open && !isWithinOpenDetailsSummary(element, current)) ||
        current.hasAttribute("hidden") || (!isSlot && !isVisibleInTabbableTree(current, isAncestor))) {
        return false;
      }
    }
    return true;
  }

  function getTabIndex(element) {
    const tabIndex = element.tabIndex;
    if (tabIndex < 0) {
      const name = nodeName(element);
      if (name === "details" || name === "audio" || name === "video" || (isHTMLElement(element) && element.isContentEditable)) return 0;
    }
    return tabIndex;
  }

  function getNamedRadioInput(element) {
    if (nodeName(element) !== "input") return null;
    return element.type === "radio" && element.name !== "" ? element : null;
  }

  function isTabbableRadio(element, candidates) {
    const input = getNamedRadioInput(element);
    if (!input) return true;
    const checkedRadio = candidates.find((candidate) => {
      const radio = getNamedRadioInput(candidate);
      return radio?.name === input.name && radio.form === input.form && radio.checked;
    });
    if (checkedRadio) return checkedRadio === input;
    return candidates.find((candidate) => {
      const radio = getNamedRadioInput(candidate);
      return radio?.name === input.name && radio.form === input.form;
    }) === input;
  }

  function getComposedChildren(container) {
    if (isHTMLElement(container) && nodeName(container) === "slot") {
      const assignedElements = container.assignedElements({ flatten: true });
      if (assignedElements.length > 0) return assignedElements;
    }
    if (isHTMLElement(container) && container.shadowRoot) return Array.from(container.shadowRoot.children);
    return Array.from(container.children);
  }

  function appendCandidates(container, list) {
    getComposedChildren(container).forEach((child) => {
      if (isFocusableCandidate(child)) list.push(child);
      appendCandidates(child, list);
    });
  }

  function appendMatchingElements(container, selector, list) {
    getComposedChildren(container).forEach((child) => {
      if (isHTMLElement(child) && child.matches(selector)) list.push(child);
      appendMatchingElements(child, selector, list);
    });
  }

  function isTabbable(element) {
    return isFocusableElement(element) && getTabIndex(element) >= 0;
  }

  function focusable(container) {
    const candidates = [];
    appendCandidates(container, candidates);
    return candidates.filter(isFocusableElement);
  }

  function tabbable(container) {
    const candidates = focusable(container);
    return candidates.filter((element) => getTabIndex(element) >= 0 && isTabbableRadio(element, candidates));
  }

  function getTabbableIn(container, dir) {
    const list = tabbable(container);
    const len = list.length;
    if (len === 0) return undefined;
    const active = activeElement(container.ownerDocument);
    const index = list.indexOf(active);
    const nextIndex = index === -1 ? (dir === 1 ? 0 : len - 1) : index + dir;
    return list[nextIndex];
  }

  function getNextTabbable(referenceElement) {
    return getTabbableIn(referenceElement.ownerDocument.body, 1) || referenceElement;
  }

  function getPreviousTabbable(referenceElement) {
    return getTabbableIn(referenceElement.ownerDocument.body, -1) || referenceElement;
  }

  function getTabbableNearElement(referenceElement, dir) {
    if (!referenceElement) return null;
    const list = tabbable(referenceElement.ownerDocument.body);
    const elementCount = list.length;
    if (elementCount === 0) return null;
    const index = list.indexOf(referenceElement);
    if (index === -1) return null;
    return list[(index + dir + elementCount) % elementCount];
  }

  function getTabbableAfterElement(referenceElement) {
    return getTabbableNearElement(referenceElement, 1);
  }

  function getTabbableBeforeElement(referenceElement) {
    return getTabbableNearElement(referenceElement, -1);
  }

  function isOutsideEvent(event, container) {
    const containerElement = container || event.currentTarget;
    const relatedTarget = event.relatedTarget;
    return !relatedTarget || !contains(containerElement, relatedTarget);
  }

  // data-tabindex is Base UI's own marker for the tabindex it took away.
  function disableFocusInside(container) {
    tabbable(container).forEach((element) => {
      element.dataset.tabindex = element.getAttribute("tabindex") || "";
      element.setAttribute("tabindex", "-1");
    });
  }

  function enableFocusInside(container) {
    const elements = [];
    appendMatchingElements(container, "[data-tabindex]", elements);
    elements.forEach((element) => {
      const tabindex = element.dataset.tabindex;
      delete element.dataset.tabindex;
      if (tabindex) element.setAttribute("tabindex", tabindex);
      else element.removeAttribute("tabindex");
    });
  }

  window.templ = window.templ || {};
  window.templ.tabbable = {
    activeElement,
    contains,
    isElementVisible,
    isTabbable,
    focusable,
    tabbable,
    getNextTabbable,
    getPreviousTabbable,
    getTabbableAfterElement,
    getTabbableBeforeElement,
    isOutsideEvent,
    disableFocusInside,
    enableFocusInside,
  };
})();

// components/baseui/use_click.js
// Port of @base-ui/react floating-ui-react/hooks/useClick.ts (1.6.0).
// Opens or closes a popup when its trigger is pressed.
//
//   const cleanup = window.templ.click.useClick(trigger, options)
//
//   event           "click" (default), "mousedown", or "mousedown-only"
//   toggle          a repeated press closes (default true), or a function
//                   read on every press
//   ignoreMouse     mouse presses do nothing (default false)
//   stickIfOpen     a popup opened by hover or focus stays on a click (default
//                   true), or a function read on every press
//   touchOpenDelay  ms before a touch opens (default 0)
//   isOpen()        whether the popup is open
//   isActiveTrigger()  whether this trigger opened it, default true
//   openEventType() the type of the event that opened it, if any
//   onOpenChange(open, event, pointerType)  requests the change
//
// With "mousedown" the open waits one frame, until the browser moved focus,
// instead of preventing the default, which would show :focus-visible.
(function () {
  "use strict";

  const TYPEABLE_SELECTOR = "input:not([type='hidden']):not([disabled]),[contenteditable]:not([contenteditable='false']),textarea:not([disabled])";
  const isMouseLikePointerType = (pointerType) => pointerType === "mouse" || pointerType === "pen";

  function useClick(trigger, options) {
    const {
      event: eventOption = "click",
      toggle = true,
      ignoreMouse = false,
      stickIfOpen = true,
      touchOpenDelay = 0,
      isOpen,
      isActiveTrigger = () => true,
      openEventType = () => null,
      onOpenChange,
    } = options;
    const read = (value) => (typeof value === "function" ? value() : value);
    let pointerType;
    let frame = 0;
    let touchOpenTimer = 0;

    function setOpenWithTouchDelay(nextOpen, event, type) {
      if (nextOpen && type === "touch" && touchOpenDelay > 0) {
        clearTimeout(touchOpenTimer);
        touchOpenTimer = setTimeout(() => onOpenChange(true, event, type), touchOpenDelay);
      } else {
        onOpenChange(nextOpen, event, type);
      }
    }

    function getNextOpen(isClickLikeOpenEvent) {
      const open = isOpen();
      // Moving between triggers always opens the newly active one.
      if (open && !isActiveTrigger()) return true;
      // A closed popup opens on the next press.
      if (!open) return true;
      // Non toggle mode never closes on a repeated press.
      if (!read(toggle)) return true;
      // A popup opened by hover or focus stays until a click like event closes it.
      const openType = openEventType();
      if (openType && read(stickIfOpen)) return !isClickLikeOpenEvent(openType);
      return false;
    }

    function onPointerDown(event) {
      pointerType = event.pointerType;
    }

    function onMouseDown(event) {
      const type = pointerType;
      if (event.button !== 0 || eventOption === "click" || (isMouseLikePointerType(type) && ignoreMouse)) return;
      const nextOpen = getNextOpen((openType) => openType === "click" || openType === "mousedown");
      // Focus is always set on typeable elements, which open at once.
      if (event.target instanceof Element && event.target.matches(TYPEABLE_SELECTOR)) {
        setOpenWithTouchDelay(nextOpen, event, type);
        return;
      }
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(() => setOpenWithTouchDelay(nextOpen, event, type));
    }

    function onClick(event) {
      if (eventOption === "mousedown-only") return;
      const type = pointerType;
      if (eventOption === "mousedown" && type) {
        pointerType = undefined;
        return;
      }
      if (isMouseLikePointerType(type) && ignoreMouse) return;
      const nextOpen = getNextOpen((openType) => ["click", "mousedown", "keydown", "keyup"].includes(openType));
      setOpenWithTouchDelay(nextOpen, event, type);
    }

    function onKeyDown() {
      pointerType = undefined;
    }

    trigger.addEventListener("pointerdown", onPointerDown);
    trigger.addEventListener("mousedown", onMouseDown);
    trigger.addEventListener("click", onClick);
    trigger.addEventListener("keydown", onKeyDown);
    return () => {
      cancelAnimationFrame(frame);
      clearTimeout(touchOpenTimer);
      trigger.removeEventListener("pointerdown", onPointerDown);
      trigger.removeEventListener("mousedown", onMouseDown);
      trigger.removeEventListener("click", onClick);
      trigger.removeEventListener("keydown", onKeyDown);
    };
  }

  window.templ = window.templ || {};
  window.templ.click = { useClick };
})();

// components/baseui/use_dismiss.js
// Port of @base-ui/react floating-ui-react/hooks/useDismiss.ts (1.6.0).
// Closes an open popup on Escape and on a press outside of it.
//
// useDismiss(options) runs while the popup is open, like the source's effect:
// call it on open and call the cleanup it returns on close.
//
//   floating           the popup element (required)
//   reference          the trigger element, or a list of them (the input for a combobox)
//   onOpenChange(open, reason, event)  requests the close, returns true when accepted
//   escapeKey          true, or a function read on every key press
//   outsidePress       true, false, or a function(event) that returns whether to dismiss
//   outsidePressEvent  "sloppy" (press), "intentional" (click), { mouse, touch }, or a function
//   referencePress     false, true, or a function: a press on the reference closes
//   bubbles            true, or { escapeKey, outsidePress }
//
// What differs from the source comes from having no React tree:
// - Base UI marks events inside its React tree, portals included. Here that tree
//   is the DOM plus the portal owners from portal.js, see withinTree.
// - The floating tree is every open useDismiss: a child is one whose popup lies
//   inside this popup's tree.
// - Options React re-reads on every render may be functions, read per event.
(function () {
  "use strict";

  const webkit = typeof CSS !== "undefined" && !!CSS.supports?.("-webkit-backdrop-filter:none");

  // Open instances, the FloatingTree pendant.
  const instances = new Set();

  // IME composition: Escape that settles a composition closes the compose
  // menu, not the popup. Safari fires compositionend before keydown.
  let isComposing = false;
  let compositionTimer = 0;
  document.addEventListener("compositionstart", () => {
    clearTimeout(compositionTimer);
    isComposing = true;
  });
  document.addEventListener("compositionend", () => {
    compositionTimer = setTimeout(() => {
      isComposing = false;
    }, webkit ? 5 : 0);
  });

  function getTarget(event) {
    return event.composedPath?.()[0] || event.target;
  }

  function contains(parent, child) {
    return !!parent && !!child && (parent === child || parent.contains(child));
  }

  // The React tree pendant: up through the DOM, and from a portaled element
  // on to the place it was declared.
  function withinTree(root, target) {
    for (let node = target; node; node = window.templ.portal.treeParent(node)) {
      if (node === root) return true;
    }
    return false;
  }

  function isLastTraversableNode(node) {
    return ["html", "body", "#document"].includes(node.nodeName.toLowerCase());
  }

  function normalizeBubbles(bubbles) {
    return {
      escapeKey: typeof bubbles === "boolean" ? bubbles : bubbles?.escapeKey ?? false,
      outsidePress: typeof bubbles === "boolean" ? bubbles : bubbles?.outsidePress ?? true,
    };
  }

  function useDismiss(options) {
    const {
      floating,
      onOpenChange,
      escapeKey = true,
      outsidePress = true,
      outsidePressEvent = "sloppy",
      referencePress = false,
    } = options;
    const { reference } = options;
    const references = (reference instanceof Element ? [reference] : [...(reference || [])]).filter(Boolean);
    const bubbles = normalizeBubbles(options.bubbles);
    const self = { floating, bubbles, active: true };

    let pressStartedInside = false;
    let pressStartPrevented = false;
    // Ignore only the very next outside click after dragging from inside to outside.
    let suppressNextOutsideClick = false;
    let currentPointerType = "";
    let touchState = null;
    let cancelDismissOnEndTimer = 0;
    let preventedPressSuppressionTimer = 0;

    const read = (value) => (typeof value === "function" ? value() : value);

    function hasBlockingChild(bubbleKey) {
      for (const other of instances) {
        if (other !== self && withinTree(floating, other.floating) && !other.bubbles[bubbleKey]) return true;
      }
      return false;
    }

    function isEventWithinOwnElements(event) {
      const target = getTarget(event);
      return contains(floating, target) || references.some((reference) => contains(reference, target));
    }

    function close(reason, event) {
      return self.active && onOpenChange(false, reason, event);
    }

    function closeOnEscapeKeyDown(event) {
      if (!self.active || !read(escapeKey) || event.key !== "Escape") return;
      if (isComposing) return;
      if (!bubbles.escapeKey && hasBlockingChild("escapeKey")) return;
      if (close("escape-key", event)) event.preventDefault();
      if (!bubbles.escapeKey) event.stopPropagation();
    }

    // The popup's own press handlers (the source's floating props).
    function markPressStartedInside(event) {
      if (event.button !== 0) return;
      // Only presses that start in the popup's DOM subtree count as inside,
      // so a nested portal does not suppress the parent's dismissal.
      if (!contains(floating, getTarget(event))) return;
      if (!pressStartedInside) {
        pressStartedInside = true;
        pressStartPrevented = false;
      }
    }

    function markInsidePressStartPrevented(event) {
      if (event.defaultPrevented && pressStartedInside) pressStartPrevented = true;
    }

    function getOutsidePressEvent() {
      const type = currentPointerType === "pen" || !currentPointerType ? "mouse" : currentPointerType;
      const value = read(outsidePressEvent);
      return typeof value === "string" ? value : value[type];
    }

    function shouldIgnoreEvent(event) {
      const mode = getOutsidePressEvent();
      return (mode === "intentional" && event.type !== "click") || (mode === "sloppy" && event.type === "click");
    }

    function suppressImmediateOutsideClickAfterPreventedStart() {
      suppressNextOutsideClick = true;
      clearTimeout(preventedPressSuppressionTimer);
      preventedPressSuppressionTimer = setTimeout(() => {
        suppressNextOutsideClick = false;
      }, 0);
    }

    function resetPressStartState() {
      pressStartedInside = false;
      pressStartPrevented = false;
    }

    function closeOnPressOutside(event) {
      if (!self.active) return;
      if (shouldIgnoreEvent(event)) {
        // A new press began outside the popup and its trigger. Clear any
        // leftover drag-out suppression so this press's click can dismiss.
        if (event.type !== "click" && !isEventWithinOwnElements(event)) {
          clearTimeout(preventedPressSuppressionTimer);
          suppressNextOutsideClick = false;
        }
        return;
      }
      const target = getTarget(event);
      // Inside the popup's tree, portaled children included.
      if (withinTree(floating, target)) return;
      // Another trigger of this popup was pressed.
      if (references.some((reference) => contains(reference, target))) return;

      // A press on a third party element injected after the popup opened:
      // its top level ancestor carries none of the data-base-ui-inert markers
      // markOthers set when it opened.
      const targetRoot = target instanceof Element ? target.getRootNode() : null;
      const isShadowRoot = typeof ShadowRoot !== "undefined" && targetRoot instanceof ShadowRoot;
      const markers = Array.from((isShadowRoot ? targetRoot : floating.ownerDocument).querySelectorAll("[data-base-ui-inert]"));
      let targetRootAncestor = target instanceof Element ? target : null;
      while (targetRootAncestor && !isLastTraversableNode(targetRootAncestor)) {
        const nextParent = targetRootAncestor.assignedSlot || targetRootAncestor.parentNode ||
          (targetRootAncestor.parentNode instanceof ShadowRoot ? targetRootAncestor.parentNode.host : null);
        if (!nextParent || isLastTraversableNode(nextParent) || !(nextParent instanceof Element)) break;
        targetRootAncestor = nextParent;
      }
      if (markers.length && target instanceof Element && !target.matches("html,body") &&
        !contains(target, floating) && markers.every((marker) => !contains(targetRootAncestor, marker))) {
        return;
      }

      // A press on a scrollbar. Touch never hits a scrollbar.
      if (target instanceof HTMLElement && !("touches" in event)) {
        const lastTraversableNode = isLastTraversableNode(target);
        const style = getComputedStyle(target);
        const scrollRe = /auto|scroll/;
        const isScrollableX = lastTraversableNode || scrollRe.test(style.overflowX);
        const isScrollableY = lastTraversableNode || scrollRe.test(style.overflowY);
        const canScrollX = isScrollableX && target.clientWidth > 0 && target.scrollWidth > target.clientWidth;
        const canScrollY = isScrollableY && target.clientHeight > 0 && target.scrollHeight > target.clientHeight;
        const isRTL = style.direction === "rtl";
        const pressedVerticalScrollbar = canScrollY &&
          (isRTL ? event.offsetX <= target.offsetWidth - target.clientWidth : event.offsetX > target.clientWidth);
        const pressedHorizontalScrollbar = canScrollX && event.offsetY > target.clientHeight;
        if (pressedVerticalScrollbar || pressedHorizontalScrollbar) return;
      }

      // In intentional mode, a press that starts inside and ends outside gets
      // one suppressed outside click.
      if (getOutsidePressEvent() === "intentional" && suppressNextOutsideClick) {
        clearTimeout(preventedPressSuppressionTimer);
        suppressNextOutsideClick = false;
        return;
      }
      if (typeof outsidePress === "function" && !outsidePress(event)) return;
      if (hasBlockingChild("outsidePress")) return;
      close("outside-press", event);
    }

    function handlePointerDown(event) {
      if (getOutsidePressEvent() !== "sloppy" || event.pointerType === "touch" || isEventWithinOwnElements(event)) return;
      closeOnPressOutside(event);
    }

    function handleTouchStart(event) {
      if (getOutsidePressEvent() !== "sloppy" || isEventWithinOwnElements(event)) return;
      const touch = event.touches[0];
      if (!touch) return;
      touchState = {
        startX: touch.clientX,
        startY: touch.clientY,
        dismissOnTouchEnd: false,
        dismissOnMouseDown: true,
      };
      clearTimeout(cancelDismissOnEndTimer);
      cancelDismissOnEndTimer = setTimeout(() => {
        if (touchState) {
          touchState.dismissOnTouchEnd = false;
          touchState.dismissOnMouseDown = false;
        }
      }, 1000);
    }

    // Runs the listener after the target's own handlers, in the bubble phase
    // of the target, like the source.
    function addTargetEventListenerOnce(event, listener) {
      const target = getTarget(event);
      if (!target) return;
      const once = () => {
        target.removeEventListener(event.type, once);
        listener(event);
      };
      target.addEventListener(event.type, once);
    }

    function handleTouchStartCapture(event) {
      currentPointerType = "touch";
      addTargetEventListenerOnce(event, handleTouchStart);
    }

    function closeOnPressOutsideCapture(event) {
      clearTimeout(cancelDismissOnEndTimer);
      if (event.type === "pointerdown") currentPointerType = event.pointerType;
      if (event.type === "mousedown" && touchState && !touchState.dismissOnMouseDown) return;
      addTargetEventListenerOnce(event, (targetEvent) => {
        if (targetEvent.type === "pointerdown") handlePointerDown(targetEvent);
        else closeOnPressOutside(targetEvent);
      });
    }

    function handlePressEndCapture(event) {
      if (!pressStartedInside) return;
      const startPrevented = pressStartPrevented;
      resetPressStartState();
      if (getOutsidePressEvent() !== "intentional") return;
      if (event.type === "pointercancel") {
        if (startPrevented) suppressImmediateOutsideClickAfterPreventedStart();
        return;
      }
      if (withinTree(floating, getTarget(event))) return;
      if (startPrevented) {
        suppressImmediateOutsideClickAfterPreventedStart();
        return;
      }
      if (typeof outsidePress === "function" && !outsidePress(event)) return;
      clearTimeout(preventedPressSuppressionTimer);
      suppressNextOutsideClick = true;
    }

    function handleTouchMove(event) {
      if (getOutsidePressEvent() !== "sloppy" || !touchState || isEventWithinOwnElements(event)) return;
      const touch = event.touches[0];
      if (!touch) return;
      const distance = Math.hypot(touch.clientX - touchState.startX, touch.clientY - touchState.startY);
      if (distance > 5) touchState.dismissOnTouchEnd = true;
      if (distance > 10) {
        closeOnPressOutside(event);
        clearTimeout(cancelDismissOnEndTimer);
        touchState = null;
      }
    }

    function handleTouchEnd(event) {
      if (getOutsidePressEvent() !== "sloppy" || !touchState || isEventWithinOwnElements(event)) return;
      if (touchState.dismissOnTouchEnd) closeOnPressOutside(event);
      clearTimeout(cancelDismissOnEndTimer);
      touchState = null;
    }

    // The reference's onPointerDown and onClick.
    function closeOnReferencePress(event) {
      if (!read(referencePress)) return;
      close("trigger-press", event);
    }

    const listeners = [
      ...references.map((reference) => [reference, "pointerdown", closeOnReferencePress]),
      ...references.map((reference) => [reference, "click", closeOnReferencePress]),
      [document, "keydown", closeOnEscapeKeyDown],
      ...references.map((reference) => [reference, "keydown", closeOnEscapeKeyDown]),
      [floating, "keydown", closeOnEscapeKeyDown],
      [floating, "pointerdown", markPressStartedInside, true],
      [floating, "mousedown", markPressStartedInside, true],
      [floating, "pointerdown", markInsidePressStartPrevented],
      [floating, "mousedown", markInsidePressStartPrevented],
    ];
    if (outsidePress !== false) {
      listeners.push(
        [document, "click", closeOnPressOutsideCapture, true],
        [document, "pointerdown", closeOnPressOutsideCapture, true],
        [document, "pointerup", handlePressEndCapture, true],
        [document, "pointercancel", handlePressEndCapture, true],
        [document, "mousedown", closeOnPressOutsideCapture, true],
        [document, "mouseup", handlePressEndCapture, true],
        [document, "touchstart", handleTouchStartCapture, true],
        [document, "touchmove", (event) => addTargetEventListenerOnce(event, handleTouchMove), true],
        [document, "touchend", (event) => addTargetEventListenerOnce(event, handleTouchEnd), true],
      );
    }
    listeners.forEach(([target, type, listener, capture]) => target.addEventListener(type, listener, !!capture));
    instances.add(self);

    return function cleanup() {
      if (!self.active) return;
      self.active = false;
      instances.delete(self);
      listeners.forEach(([target, type, listener, capture]) => target.removeEventListener(type, listener, !!capture));
      clearTimeout(cancelDismissOnEndTimer);
      clearTimeout(preventedPressSuppressionTimer);
    };
  }

  window.templ = window.templ || {};
  window.templ.dismiss = { useDismiss };
})();

// components/baseui/use_transition_status.js
// Port of @base-ui/react internals/useTransitionStatus.ts with
// useOpenChangeComplete.ts and useAnimationsFinished.ts (1.6.0).
//
// The popup and its backdrop render the status: data-open or data-closed,
// plus data-starting-style for the first frame of an open and
// data-ending-style while it closes. The positioner and the arrow render only
// data-open or data-closed (popupStateMapping), and the positioner has its
// transitions off while starting (getDisabledMountTransitionStyles) and no
// pointer events while closed (usePositioner's inert). The close
// completes once every animation on the animated element finished, which is
// when Base UI unmounts. Opening again before that cancels the pending close.
//
// status is an array of the parts that render the status, or
// { parts, positioner, stateParts, styleParts }. parts[0] carries the pending
// status; styleParts render only the starting and ending style (the
// collapsible trigger's mapping).
(function () {
  "use strict";

  function set(parts, name, present) {
    parts.forEach((part) => part?.toggleAttribute(name, present));
  }

  function resolve(status) {
    const { parts, positioner = null, stateParts = [], styleParts = [] } = Array.isArray(status) ? { parts: status } : status;
    const own = parts.filter(Boolean);
    return {
      parts: [...own, ...styleParts.filter(Boolean)],
      open: [...own, positioner, ...stateParts].filter(Boolean),
      positioner,
    };
  }

  // usePositioner's inert: a closed positioner takes no pointer events, also
  // during its exit animation. Opening only takes back that none, so a value
  // safePolygon set stays, as React leaves a style it did not render.
  function setOpen(status, isOpen) {
    set(status.open, "data-open", isOpen);
    set(status.open, "data-closed", !isOpen);
    const style = status.positioner?.style;
    if (!style) return;
    if (!isOpen) style.pointerEvents = "none";
    else if (style.pointerEvents === "none") style.pointerEvents = "";
  }

  // useAnimationsFinished: runs fn once every animation on the element
  // finished, as long as isCurrent() still holds. An animation aborted
  // because a property it depends on changed may be followed by a new one,
  // so check again before running fn.
  function whenAnimationsFinish(animated, isCurrent, fn) {
    const exec = () => {
      if (!isCurrent()) return;
      if (typeof animated?.getAnimations !== "function") return fn();
      Promise.all(animated.getAnimations().map((animation) => animation.finished)).then(() => {
        if (isCurrent()) fn();
      }, () => {
        const running = animated.getAnimations().some((a) => a.pending || a.playState !== "finished");
        if (running) exec();
        else if (isCurrent()) fn();
      });
    };
    // One frame, so the new style's animations are registered.
    requestAnimationFrame(exec);
  }

  // onComplete, when given, runs once the open animations on animated
  // finished (useOpenChangeComplete with open true).
  function open(statusParam, animated, onComplete) {
    const status = resolve(statusParam);
    const { parts, positioner } = status;
    const main = parts[0];
    const token = {};
    main._templTransition = token;
    set(parts, "data-ending-style", false);
    setOpen(status, true);
    set(parts, "data-starting-style", true);
    if (positioner) positioner.style.transition = "none";
    // Compute the starting style once, so transitions start from it.
    void main.offsetWidth;
    requestAnimationFrame(() => {
      if (main._templTransition !== token) return;
      set(parts, "data-starting-style", false);
      if (positioner) positioner.style.transition = "";
      if (onComplete) whenAnimationsFinish(animated, () => main._templTransition === token, onComplete);
    });
  }

  // deferEnding sets data-ending-style one frame after data-closed, Base UI's
  // deferEndingState, and calls onEnding then. onEnding returning false
  // completes the close at once.
  function close(statusParam, animated, onComplete, { deferEnding = false, onEnding } = {}) {
    const status = resolve(statusParam);
    const { parts, positioner } = status;
    const main = parts[0];
    const token = {};
    main._templTransition = token;
    set(parts, "data-starting-style", false);
    if (positioner) positioner.style.transition = "";
    setOpen(status, false);

    const done = () => {
      if (main._templTransition !== token) return;
      main._templTransition = null;
      set(parts, "data-ending-style", false);
      onComplete?.();
    };
    const ending = () => {
      set(parts, "data-ending-style", true);
      if (onEnding?.() === false) return done();
      whenAnimationsFinish(animated, () => main._templTransition === token, done);
    };
    if (!deferEnding) return ending();
    requestAnimationFrame(() => {
      if (main._templTransition === token) ending();
    });
  }

  // Sets the state without a transition, like an unmount without exit
  // animation, and cancels one in flight.
  function reset(statusParam, isOpen) {
    const status = resolve(statusParam);
    status.parts[0]._templTransition = null;
    set(status.parts, "data-starting-style", false);
    set(status.parts, "data-ending-style", false);
    if (status.positioner) status.positioner.style.transition = "";
    setOpen(status, isOpen);
  }

  // Whether a close is in flight, Base UI's transitionStatus === "ending".
  function isEnding(element) {
    return !!element?.hasAttribute("data-ending-style");
  }

  // useAnimationsFinished on its own, for parts whose status another store
  // drives, like the toast manager.
  function animationsFinished(element, fn) {
    whenAnimationsFinish(element, () => element.isConnected, fn);
  }

  window.templ = window.templ || {};
  window.templ.transition = { open, close, reset, isEnding, animationsFinished };
})();

// components/dialog/dialog.js
(function () {
  "use strict";

  // Vanilla port of Base UI's Dialog (packages/react/src/dialog): the portal
  // node, backdrop and popup are SSRd divs. The script wires the blocks in
  // components/baseui (portal, transition status, dismiss, focus manager,
  // scroll lock) the way DialogRoot and DialogPopup use them. "Unmount" is
  // the portal node getting [hidden] again.

  // ----- registry ------------------------------------------------------------

  // Popup element -> per-dialog state. The open stack orders open dialogs by
  // open time (last = topmost), like Base UI's nested dialog counts.
  const dialogs = new Map();
  const openStack = [];

  // A dialog popup (Dialog, Sheet, AlertDialog) is the dialog or alertdialog
  // element carrying Base UI's modal prop; the popover popup has role dialog
  // too but no modal prop. Its parent is the portal node (data-base-ui-portal)
  // that holds the backdrop and the popup.
  const POPUP = '[role="dialog"][data-templ-modal], [role="alertdialog"][data-templ-modal]';
  // Base UI's Dialog.Backdrop renders role="presentation".
  const BACKDROP = ':scope > [role="presentation"]';

  function getDialog(target) {
    if (!target) return null;
    if (typeof target === "string") {
      const el = document.getElementById(target);
      return el && el.matches(POPUP) ? el : null;
    }
    if (target.matches?.(POPUP)) return target;
    return target.closest?.(POPUP) || null;
  }

  function stateOf(target) {
    const popup = getDialog(target);
    return popup ? dialogs.get(popup) : null;
  }

  function dialogFor(element) {
    // Dialog.Close links through context in Base UI; its port marker carries
    // the dialog id when the close sits outside the popup.
    const id =
      element.getAttribute("data-templ-controls") || element.getAttribute("data-templ-dialog-close");
    if (id) return getDialog(id);
    return getDialog(element);
  }

  function triggersFor(popup) {
    if (!popup.id) return [];
    return document.querySelectorAll(
      '[data-base-ui-click-trigger][data-templ-controls="' + popup.id + '"]',
    );
  }

  function isModal(state) {
    return state.popup.getAttribute("data-templ-modal") !== "false";
  }

  // ----- interaction type ----------------------------------------------------

  // FloatingFocusManager tracks the last pointer/keyboard interaction to pick
  // touch initial focus and keyboard-visible return focus.
  let lastInteractionType = "";
  document.addEventListener(
    "pointerdown",
    (event) => {
      lastInteractionType = event.pointerType || "mouse";
    },
    true,
  );
  document.addEventListener(
    "keydown",
    () => {
      lastInteractionType = "keyboard";
    },
    true,
  );

  // ----- aria wiring (useDialogTitle/-Description registration) --------------

  function wireAria(state) {
    const popup = state.popup;
    const title = popup.querySelector("[data-templ-dialog-title]");
    if (title) {
      if (!title.id) title.id = popup.id + "-title";
      popup.setAttribute("aria-labelledby", title.id);
    } else {
      popup.removeAttribute("aria-labelledby");
    }
    const description = popup.querySelector("[data-templ-dialog-description]");
    if (description) {
      if (!description.id) description.id = popup.id + "-description";
      popup.setAttribute("aria-describedby", description.id);
    } else {
      popup.removeAttribute("aria-describedby");
    }
  }

  // ----- transition lifecycle ------------------------------------------------

  // The parts that render the transition status, popup first.
  function partsOf(state) {
    return [state.popup, state.backdrop];
  }

  // ----- nested dialog bookkeeping ------------------------------------------

  // A dialog is nested when its hidden portal node was SSRd inside another
  // dialog's content — the DOM pendant of Base UI's parent DialogRootContext. The
  // relation is recorded at registration (see ensureDialog); parentOf resolves
  // it to the parent's live state.
  function parentOf(state) {
    const parentId = state.root._templParent;
    return parentId ? stateOf(parentId) : null;
  }

  function nestedOpenCount(state) {
    return openStack.filter((other) => {
      for (let p = parentOf(other); p; p = parentOf(p)) {
        if (p === state) return true;
      }
      return false;
    }).length;
  }

  function updateNestedAttributes() {
    openStack.forEach((state) => {
      const count = nestedOpenCount(state);
      state.popup.style.setProperty("--nested-dialogs", String(count));
      state.popup.toggleAttribute("data-nested-dialog-open", count > 0);
    });
  }

  function isTopmost(state) {
    return nestedOpenCount(state) === 0;
  }

  // ----- open / close --------------------------------------------------------

  // DialogTrigger renders aria-controls while the popup is open
  // (triggerPopupId).
  function updateTriggers(state, isOpen) {
    triggersFor(state.popup).forEach((trigger) => {
      trigger.setAttribute("aria-expanded", isOpen ? "true" : "false");
      trigger.toggleAttribute("data-popup-open", isOpen);
      if (isOpen) trigger.setAttribute("aria-controls", state.popup.id);
      else trigger.removeAttribute("aria-controls");
    });
  }

  // DialogPortal renders an InternalBackdrop for a modal dialog while it is
  // mounted: fixed over the viewport, inert while closing, and useDismiss
  // treats it as a backdrop.
  function createInternalBackdrop() {
    const backdrop = document.createElement("div");
    backdrop.setAttribute("role", "presentation");
    backdrop.setAttribute("data-base-ui-inert", "");
    backdrop.style.cssText = "position:fixed;inset:0;user-select:none;-webkit-user-select:none";
    return backdrop;
  }

  function openDialog(target, trigger) {
    const state = stateOf(target);
    if (!state || state.open) return;

    const popup = state.popup;
    state.openType = trigger ? lastInteractionType || "mouse" : null;
    state.trigger =
      trigger && trigger instanceof Element ? trigger : triggersFor(popup)[0] || null;

    state.open = true;
    openStack.push(state);
    updateNestedAttributes();
    startDismiss(state);

    window.templ.portal.render(state.root);
    state.root.hidden = false;
    if (isModal(state)) {
      state.internalBackdrop ??= createInternalBackdrop();
      state.internalBackdrop.inert = false;
      state.root.prepend(state.internalBackdrop);
    }

    wireAria(state);

    window.templ.transition.open(partsOf(state));

    if (isModal(state)) {
      state.releaseScroll = window.templ.scrollLock.acquire(popup);
    }

    updateTriggers(state, true);

    // DialogPopup's FloatingFocusManager, mounted until the exit animation
    // finished. Opened by touch the popup takes focus, so the virtual
    // keyboard stays closed (createDefaultInitialFocus).
    if (state.focus) {
      state.focus.open();
    } else {
      state.focus = window.templ.focusManager.useFloatingFocusManager({
        floating: popup,
        reference: state.trigger,
        triggers: triggersFor(popup),
        modal: isModal(state),
        openInteractionType: state.openType,
        initialFocus: (interactionType) => (interactionType === "touch" ? popup : true),
        restoreFocus: "popup",
        closeOnFocusOut: !popup.hasAttribute("data-templ-disable-pointer-dismissal"),
        onOpenChange: (open) => requestOpenChange(popup, open),
      });
    }
  }

  function closeDialog(target) {
    const state = stateOf(target);
    if (!state || !state.open) return;

    const popup = state.popup;
    state.open = false;
    stopDismiss(state);
    const index = openStack.indexOf(state);
    if (index !== -1) openStack.splice(index, 1);
    updateNestedAttributes();

    // Base UI order on open=false: the transition status flips to ending,
    // aria-hidden marking and the scroll lock release immediately, the
    // popup unmounts (and focus returns) once the exit animation finishes.
    state.focus?.close();
    if (state.internalBackdrop) state.internalBackdrop.inert = true;
    window.templ.transition.close(partsOf(state), popup, () => {
      state.root.hidden = true;
      popup.style.removeProperty("--nested-dialogs");
      popup.removeAttribute("data-nested-dialog-open");
      // Unmounting the focus manager returns focus.
      state.focus?.unmount();
      state.focus = null;
      state.internalBackdrop?.remove();
      // onOpenChangeComplete(false) pendant: fires once the exit animation
      // finished and the dialog unmounted (command.js resets its palette on
      // this).
      popup.dispatchEvent(new CustomEvent("dialog-close", { bubbles: true }));
    });
    state.releaseScroll?.();
    state.releaseScroll = null;
    updateTriggers(state, false);
  }

  function isDialogOpen(target) {
    return stateOf(target)?.open || false;
  }

  function requestOpenChange(target, nextOpen, trigger) {
    const state = stateOf(target);
    if (!state || state.open === nextOpen) return false;
    const accepted = state.popup.dispatchEvent(
      new CustomEvent("dialog-open-change", {
        bubbles: true,
        cancelable: true,
        detail: { open: nextOpen },
      }),
    );
    if (!accepted || state.popup.hasAttribute("data-templ-open")) return false;
    if (nextOpen) openDialog(state.popup, trigger);
    else closeDialog(state.popup);
    return true;
  }

  function toggleDialog(target, trigger) {
    requestOpenChange(target, !isDialogOpen(target), trigger);
  }

  // ----- dismissal (useDismiss, options from useDialogRoot) ------------------

  function startDismiss(state) {
    const popup = state.popup;
    state.dismiss = window.templ.dismiss.useDismiss({
      floating: popup,
      reference: triggersFor(popup),
      // A nested open dialog blocks its parent.
      escapeKey: () => isTopmost(state),
      // With a backdrop the dismissal waits for the click, so a press that
      // starts inside and is released over the backdrop never dismisses.
      // Modal is a boolean here, Base UI's "trap-focus" mode does not exist.
      outsidePressEvent: () => (state.internalBackdrop?.isConnected || state.backdrop) ? "intentional" : { mouse: "intentional", touch: "sloppy" },
      outsidePress(event) {
        if ("button" in event && event.button !== 0) return false;
        if ("touches" in event && event.touches.length !== 1) return false;
        if (!isTopmost(state) || popup.hasAttribute("data-templ-disable-pointer-dismissal")) return false;
        // A modal dialog closes only on its own backdrop, which supports
        // several modal dialogs that are not nested.
        if (!isModal(state)) return true;
        const target = event.target;
        const internalBackdrop = state.internalBackdrop?.isConnected ? state.internalBackdrop : null;
        if (!state.backdrop && !internalBackdrop) return true;
        return target === state.backdrop || target === internalBackdrop ||
          (target.contains(popup) && !target.hasAttribute("data-base-ui-portal"));
      },
      onOpenChange: (open) => requestOpenChange(popup, open),
    });
  }

  function stopDismiss(state) {
    state.dismiss?.();
    state.dismiss = null;
  }

  // ----- initialization ------------------------------------------------------

  function ensureDialog(popup) {
    if (!popup || dialogs.has(popup)) return dialogs.get(popup) || null;
    const root = popup.parentElement;

    const parentPopup = root.parentElement?.closest(POPUP);
    if (parentPopup?.id) root._templParent = parentPopup.id;

    const state = {
      root,
      popup,
      backdrop: root.querySelector(BACKDROP),
      open: false,
      trigger: null,
      openType: null,
      focus: null,
      internalBackdrop: null,
      releaseScroll: null,
    };
    dialogs.set(popup, state);

    // A nested dialog renders no backdrop in Base UI (DialogBackdrop's
    // enabled: !nested); the parent's backdrop keeps covering the page.
    if (root._templParent) {
      popup.setAttribute("data-nested", "");
      if (state.backdrop) state.backdrop.hidden = true;
    }

    wireAria(state);
    return state;
  }

  // Fully retire a dialog: undo aria-hidden marking, release the scroll
  // lock and remove the portaled DOM. Used when an htmx/datastar swap
  // removed the dialog's source from the page or replaced it with a fresh
  // hidden portal node.
  function destroyDialog(popup) {
    const state = dialogs.get(popup);
    if (!state) {
      popup.parentElement?.remove();
      return;
    }
    window.templ.transition.reset(partsOf(state), false);
    stopDismiss(state);
    state.focus?.unmount();
    state.focus = null;
    const index = openStack.indexOf(state);
    if (index !== -1) openStack.splice(index, 1);
    const wasOpen = state.open;
    state.open = false;
    updateNestedAttributes();
    state.releaseScroll?.();
    state.releaseScroll = null;
    if (wasOpen) popup.dispatchEvent(new CustomEvent("dialog-close", { bubbles: true }));
    window.templ.portal.remove(state.root);
    dialogs.delete(popup);
  }

  // DialogTrigger's useClick with its default click event. Base UI's
  // DialogTrigger identifier is shared with PopoverTrigger and DrawerTrigger;
  // only triggers whose data-templ-controls names a dialog popup are ours.
  window.templ.lifecycle.register("[data-base-ui-click-trigger][data-templ-controls]", {
    init(trigger) {
      if (!dialogFor(trigger)) return;
      trigger._templDialogClick = window.templ.click.useClick(trigger, {
        isOpen: () => isDialogOpen(dialogFor(trigger)),
        onOpenChange: (nextOpen) => requestOpenChange(dialogFor(trigger), nextOpen, trigger),
      });
    },
    destroy(trigger) {
      trigger._templDialogClick?.();
      trigger._templDialogClick = null;
    },
  });

  document.addEventListener("click", (event) => {
    if (!(event.target instanceof Element)) return;
    const closeButton = event.target.closest("[data-templ-dialog-close]");
    if (closeButton) {
      requestOpenChange(dialogFor(closeButton), false);
    }
  });

  // A dialog lives as long as its SSR declaration site (the portal owner)
  // stays in the document, including trigger-less programmatic dialogs.
  // Unmounting retires it: aria-hidden marking, scroll lock, portaled DOM.
  window.templ.lifecycle.register(POPUP, {
    init(popup) {
      if (!ensureDialog(popup)) return;
      // Server-side open state (Base UI open or defaultOpen).
      if (popup.getAttribute("data-templ-open") === "true" || popup.hasAttribute("data-templ-default-open")) {
        openDialog(popup);
      }
    },
    destroy: destroyDialog,
  });

  window.templ = window.templ || {};
  window.templ.dialog = {
    open: openDialog,
    close: closeDialog,
    toggle: toggleDialog,
    isOpen: isDialogOpen,
  };
})();

