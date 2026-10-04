/* ==========================================================
   APP.JS: automatic animation and motion for every page
   No inline scripts or styles, so it works under your CSP.
========================================================== */

(function () {
    "use strict";

    document.documentElement.classList.add("js");

    var reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    /* ---------- What gets animated ---------- */

    // Elements inside these are never animated
    var SKIP_INSIDE =
        ".hero-in, .reveal, .anim, [data-anim-done], .site-header, .site-footer, " +
        ".modal, .dropdown-menu, .offcanvas, .toast, .chat-messages, [data-no-anim]";

    // If an element already contains animated children, leave it alone
    var HAS_ANIMATED_INSIDE = ".anim, .hero-in, .reveal";

    // These grow in (zoom). Everything else rises up.
    var ZOOM = ".card, .alert, img, table, .auth-card, .feature-card, .step-card, .cta-card";

    // Everything inside <main> that should animate
    var BLOCKS =
        "main h1, main h2, main h3, main h4, main h5, main h6, main p, " +
        "main form, main .card, main .alert, main table, main img, " +
        "main .btn, main .list-group";

    function ready(fn) {
        if (document.readyState === "loading") {
            document.addEventListener("DOMContentLoaded", fn);
        } else {
            fn();
        }
    }

    function isCol(el) {
        return Array.prototype.some.call(el.classList, function (name) {
            return name === "col" || name.indexOf("col-") === 0;
        });
    }

    /* ---------- Scroll reveal ---------- */

    var observer = null;

    function shouldSkip(el) {
        return el.closest(SKIP_INSIDE) !== null || el.querySelector(HAS_ANIMATED_INSIDE) !== null;
    }

    function tag(el, variant, order) {
        if (shouldSkip(el)) return;

        var parent = el.parentElement;
        var n = order.get(parent) || 0;
        order.set(parent, n + 1);

        el.setAttribute("data-anim-done", "1");
        el.classList.add("anim", "anim-" + variant);
        el.style.setProperty("--anim-delay", Math.min(n, 6) * 90 + "ms");

        observer.observe(el);
    }

    function tagElements() {
        var order = new Map();

        // 1) Bootstrap columns
        document.querySelectorAll("main .row").forEach(function (row) {
            var cols = Array.prototype.filter.call(row.children, isCol);
            cols.forEach(function (col, i) {
                var variant = cols.length === 2 ? (i === 0 ? "left" : "right") : "up";
                tag(col, variant, order);
            });
        });

        // 2) Everything else
        document.querySelectorAll(BLOCKS).forEach(function (el) {
            tag(el, el.matches(ZOOM) ? "zoom" : "up", order);
        });
    }

    function finish(el, onEnd) {
        el.removeEventListener("transitionend", onEnd);
        el.classList.remove("anim", "anim-up", "anim-left", "anim-right", "anim-zoom", "is-visible");
        el.style.removeProperty("--anim-delay");
    }

    function show(el) {
        el.classList.add("is-visible");

        // Remove animation classes afterwards so the element's own
        // hover effects (cards, buttons) work normally again.
        function onEnd(e) {
            if (e.target === el && e.propertyName === "opacity") finish(el, onEnd);
        }
        el.addEventListener("transitionend", onEnd);
        setTimeout(function () { finish(el, onEnd); }, 2500);
    }

    /* ---------- Number counters: <h3 data-count="1200" data-suffix="+">0</h3> ---------- */

    function runCounter(el) {
        var end = parseFloat(el.getAttribute("data-count"));
        if (isNaN(end)) return;

        var suffix = el.getAttribute("data-suffix") || "";
        var duration = 1400;
        var start = null;

        function step(time) {
            if (start === null) start = time;
            var p = Math.min((time - start) / duration, 1);
            var eased = 1 - Math.pow(1 - p, 3);
            el.textContent = Math.round(end * eased).toLocaleString() + suffix;
            if (p < 1) requestAnimationFrame(step);
        }
        requestAnimationFrame(step);
    }

    /* ---------- Scroll effects: progress bar, header shadow, back-to-top ---------- */

    function setupScrollEffects() {
        var header = document.querySelector(".site-header");

        var bar = document.createElement("div");
        bar.className = "scroll-progress";
        bar.setAttribute("aria-hidden", "true");

        var topBtn = document.createElement("button");
        topBtn.type = "button";
        topBtn.className = "to-top";
        topBtn.setAttribute("aria-label", "Back to top");
        topBtn.textContent = "\u2191";

        document.body.appendChild(bar);
        document.body.appendChild(topBtn);

        topBtn.addEventListener("click", function () {
            window.scrollTo({ top: 0, behavior: "smooth" });
        });

        var ticking = false;

        function update() {
            var y = window.pageYOffset;
            var max = document.documentElement.scrollHeight - window.innerHeight;

            bar.style.setProperty("--progress", String(max > 0 ? y / max : 0));
            if (header) header.classList.toggle("is-scrolled", y > 10);
            topBtn.classList.toggle("is-shown", y > 500);
            ticking = false;
        }

        function onScroll() {
            if (ticking) return;
            ticking = true;
            requestAnimationFrame(update);
        }

        window.addEventListener("scroll", onScroll, { passive: true });
        window.addEventListener("resize", onScroll);
        update();
    }

    /* ---------- Button ripple ---------- */

    function setupRipple() {
        document.addEventListener("click", function (e) {
            var btn = e.target.closest(".btn, .site-btn");
            if (!btn || btn.disabled || btn.classList.contains("disabled")) return;

            var rect = btn.getBoundingClientRect();
            var size = Math.max(rect.width, rect.height) * 2;

            // keyboard clicks have no pointer position, so use the center
            var x = e.detail === 0 ? rect.width / 2 : e.clientX - rect.left;
            var y = e.detail === 0 ? rect.height / 2 : e.clientY - rect.top;

            var ripple = document.createElement("span");
            ripple.className = "ripple";
            ripple.style.width = size + "px";
            ripple.style.height = size + "px";
            ripple.style.left = x - size / 2 + "px";
            ripple.style.top = y - size / 2 + "px";

            btn.appendChild(ripple);
            ripple.addEventListener("animationend", function () { ripple.remove(); });
        });
    }

    /* ---------- Shake invalid form fields ---------- */

    function setupInvalidShake() {
        document.addEventListener("invalid", function (e) {
            var el = e.target;
            el.classList.remove("shake");
            void el.offsetWidth; // restart the animation
            el.classList.add("shake");
            el.addEventListener("animationend", function () {
                el.classList.remove("shake");
            }, { once: true });
        }, true);
    }

    /* ---------- Page leave transition ---------- */

    function setupPageTransition() {
        document.addEventListener("click", function (e) {
            if (e.defaultPrevented || e.button !== 0 ||
                e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;

            var link = e.target.closest("a[href]");
            if (!link) return;
            if ((link.target && link.target !== "_self") || link.hasAttribute("download")) return;

            var url;
            try { url = new URL(link.href, window.location.href); } catch (err) { return; }

            if (url.origin !== window.location.origin) return;

            // same page (for example only the #hash changes): no transition
            if (url.pathname === window.location.pathname &&
                url.search === window.location.search) return;

            e.preventDefault();
            document.body.classList.add("page-leaving");
            setTimeout(function () { window.location.href = url.href; }, 180);
        });

        // coming back with the browser Back button
        window.addEventListener("pageshow", function () {
            document.body.classList.remove("page-leaving");
        });
    }

    /* ---------- Start ---------- */

    ready(function () {
        var reveals = document.querySelectorAll(".reveal");
        var counters = document.querySelectorAll("[data-count]");

        // No motion: show everything immediately
        if (reduceMotion || !("IntersectionObserver" in window)) {
            reveals.forEach(function (el) { el.classList.add("is-visible"); });
            counters.forEach(function (el) {
                el.textContent =
                    Number(el.getAttribute("data-count")).toLocaleString() +
                    (el.getAttribute("data-suffix") || "");
            });
            return;
        }

        observer = new IntersectionObserver(function (entries) {
            entries.forEach(function (entry) {
                if (!entry.isIntersecting) return;

                var el = entry.target;
                observer.unobserve(el);

                if (el.hasAttribute("data-count")) runCounter(el);

                if (el.classList.contains("anim")) {
                    show(el);
                } else if (el.classList.contains("reveal")) {
                    el.classList.add("is-visible");
                }
            });
        }, { threshold: 0.12, rootMargin: "0px 0px -40px 0px" });

        tagElements();
        reveals.forEach(function (el) { observer.observe(el); });
        counters.forEach(function (el) { observer.observe(el); });

        setupScrollEffects();
        setupRipple();
        setupInvalidShake();
        setupPageTransition();

        // Call window.SiteMotion.refresh() after you add new content with JavaScript
        window.SiteMotion = { refresh: tagElements };
    });
})();