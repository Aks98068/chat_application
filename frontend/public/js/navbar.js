document.addEventListener("DOMContentLoaded", () => {


const navbar = document.getElementById("main-navbar");
const menuButton = document.getElementById("mobile-menu-button");
const mobileMenu = document.getElementById("mobile-menu");

const openIcon = document.getElementById("menu-open-icon");
const closeIcon = document.getElementById("menu-close-icon");

/*
 * ==========================================
 * Navbar scroll effect
 * ==========================================
 */

const updateNavbar = () => {

    if (!navbar) {
        return;
    }

    if (window.scrollY > 20) {

        navbar.classList.add(
            "bg-[#F7F4EE]",
            "shadow-sm"
        );

        navbar.classList.remove(
            "bg-[#F7F4EE]/90"
        );

    } else {

        navbar.classList.remove(
            "bg-[#F7F4EE]",
            "shadow-sm"
        );

        navbar.classList.add(
            "bg-[#F7F4EE]/90"
        );
    }
};

updateNavbar();

window.addEventListener(
    "scroll",
    updateNavbar,
    { passive: true }
);


/*
 * ==========================================
 * Mobile menu
 * ==========================================
 */

let menuOpen = false;

const setMenuState = (open) => {

    menuOpen = open;

    if (!mobileMenu || !menuButton) {
        return;
    }

    if (open) {

        mobileMenu.classList.remove("hidden");

        openIcon?.classList.add("hidden");
        closeIcon?.classList.remove("hidden");

        menuButton.setAttribute(
            "aria-expanded",
            "true"
        );

        menuButton.setAttribute(
            "aria-label",
            "Close menu"
        );

    } else {

        mobileMenu.classList.add("hidden");

        openIcon?.classList.remove("hidden");
        closeIcon?.classList.add("hidden");

        menuButton.setAttribute(
            "aria-expanded",
            "false"
        );

        menuButton.setAttribute(
            "aria-label",
            "Open menu"
        );
    }
};


menuButton?.addEventListener("click", () => {
    setMenuState(!menuOpen);
});


/*
 * ==========================================
 * Close menu after navigation
 * ==========================================
 */

document
    .querySelectorAll("[data-mobile-nav-link]")
    .forEach((link) => {

        link.addEventListener("click", () => {
            setMenuState(false);
        });

    });


/*
 * ==========================================
 * Active navigation link
 * ==========================================
 */

const currentPath = window.location.pathname;

document
    .querySelectorAll("[data-nav-link]")
    .forEach((link) => {

        const href = link.getAttribute("href");

        if (!href) {
            return;
        }

        const linkPath = href.split("#")[0];

        const isHome =
            linkPath === "/" &&
            currentPath === "/";

        const isActive =
            linkPath !== "/" &&
            currentPath.startsWith(linkPath);

        if (isHome || isActive) {

            link.classList.remove("text-slate-600");
            link.classList.add(
                "text-orange-500"
            );

            const indicator =
                link.querySelector(
                    "[data-nav-indicator]"
                );

            indicator?.classList.remove("hidden");
        }

    });


/*
 * ==========================================
 * Mobile active link
 * ==========================================
 */

document
    .querySelectorAll("[data-mobile-nav-link]")
    .forEach((link) => {

        const href = link.getAttribute("href");

        if (!href) {
            return;
        }

        const linkPath = href.split("#")[0];

        const isHome =
            linkPath === "/" &&
            currentPath === "/";

        const isActive =
            linkPath !== "/" &&
            currentPath.startsWith(linkPath);

        if (isHome || isActive) {

            link.classList.remove(
                "text-slate-700"
            );

            link.classList.add(
                "bg-orange-50",
                "text-orange-500"
            );
        }

    });


/*
 * ==========================================
 * Close menu when clicking outside
 * ==========================================
 */

document.addEventListener("click", (event) => {

    if (!menuOpen) {
        return;
    }

    const target = event.target;

    if (
        target instanceof Node &&
        !mobileMenu?.contains(target) &&
        !menuButton?.contains(target)
    ) {
        setMenuState(false);
    }

});


/*
 * ==========================================
 * Close mobile menu with Escape
 * ==========================================
 */

document.addEventListener("keydown", (event) => {

    if (event.key === "Escape" && menuOpen) {
        setMenuState(false);
    }

});


/*
 * ==========================================
 * Reset mobile menu when resizing
 * ==========================================
 */

window.addEventListener("resize", () => {

    if (window.innerWidth >= 1024 && menuOpen) {
        setMenuState(false);
    }

});


});
