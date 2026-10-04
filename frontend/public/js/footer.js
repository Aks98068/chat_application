
document.addEventListener("DOMContentLoaded", function () {
    const emailInput = document.getElementById("footer-email");
    const emailButton = document.getElementById("footer-email-button");

    if (!emailInput || !emailButton) {
        return;
    }

    function handleContact() {
        const email = emailInput.value.trim();

        if (!email) {
            alert("Please enter your email address.");
            emailInput.focus();
            return;
        }

        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

        if (!emailRegex.test(email)) {
            alert("Please enter a valid email address.");
            emailInput.focus();
            return;
        }

        const subject = encodeURIComponent(
            "Let's Secure Something - Portfolio Contact"
        );

        const body = encodeURIComponent(
            `Hello Abhishekh,

I found your portfolio and would like to get in touch.

My email: ${email}

Thank you.`
        );

        window.location.href =
            `mailto:abishek98071a@gmail.com?subject=${subject}&body=${body}`;
    }

    emailButton.addEventListener("click", handleContact);

    emailInput.addEventListener("keydown", function (event) {
        if (event.key === "Enter") {
            handleContact();
        }
    });
});

