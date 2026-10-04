
"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  CheckCircle2,
  Loader2,
  Mail,
  XCircle,
} from "lucide-react";

import { apiRequest } from "@/lib/api";

type VerifyEmailResponse = {
  message?: string;
};

type VerifyStatus =
  | "waiting"
  | "verifying"
  | "success"
  | "error";

export default function VerifyEmailPage() {
  const [status, setStatus] =
    useState<VerifyStatus>("waiting");

  const [message, setMessage] = useState("");

  useEffect(() => {
    const token = new URLSearchParams(
      window.location.search
    ).get("token");

    if (!token) {
      setStatus("error");
      setMessage(
        "Verification token is missing or invalid."
      );
      return;
    }

    let cancelled = false;

    async function verifyEmail() {
      if (cancelled) {
        return;
      }

      setStatus("verifying");
      setMessage("");

      try {
        console.log(
          "Sending email verification request..."
        );

        const response =
          await apiRequest<VerifyEmailResponse>(
            "/api/v1/auth/verify-email",
            {
              method: "POST",
              body: JSON.stringify({
                token,
              }),
            }
          );

        if (cancelled) {
          return;
        }

        console.log(
          "Email verification successful:",
          response
        );

        setMessage(
          response.message ||
            "Your email has been successfully verified."
        );

        setStatus("success");
      } catch (error) {
        if (cancelled) {
          return;
        }

        console.error(
          "Email verification failed:",
          error
        );

        setMessage(
          error instanceof Error
            ? error.message
            : "Unable to verify your email."
        );

        setStatus("error");
      }
    }

    verifyEmail();

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main className="sh-auth-page">
      {/* ======================================================
          BRAND PANEL
      ====================================================== */}

      <section className="sh-auth-brand">
        <div className="sh-auth-brand-content">
          <span className="sh-auth-brand-badge">
            Email verification
          </span>

          <h1 className="sh-auth-brand-title">
            One more step.
            <br />
            Verify your email.
          </h1>

          <p className="sh-auth-brand-description">
            Verify your email address to activate your
            SocialHub account and continue to your
            workspace.
          </p>

          <div className="sh-auth-brand-points">
            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                <CheckCircle2 size={16} />
              </span>

              Secure account activation
            </div>

            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                <CheckCircle2 size={16} />
              </span>

              Protect your account
            </div>

            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                <CheckCircle2 size={16} />
              </span>

              Access your SocialHub workspace
            </div>
          </div>
        </div>
      </section>

      {/* ======================================================
          VERIFICATION PANEL
      ====================================================== */}

      <section className="sh-auth-panel">
        <div className="sh-auth-form-wrap">

          {/* Logo */}

          <Link
            href="/"
            className="sh-auth-logo"
            aria-label="SocialHub home"
          >
            <span className="sh-auth-logo-mark">
              S
            </span>

            <span>SocialHub</span>
          </Link>

          {/* ==================================================
              VERIFYING
          ================================================== */}

          {status === "verifying" && (
            <>
              <div
                style={{
                  width: "56px",
                  height: "56px",
                  borderRadius: "16px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  background: "var(--primary-soft)",
                  color: "var(--primary)",
                  marginBottom: "20px",
                }}
              >
                <Loader2
                  size={28}
                  className="animate-spin"
                />
              </div>

              <h2 className="sh-auth-heading">
                Verifying your email
              </h2>

              <p className="sh-auth-subtitle">
                Please wait while we verify your
                email address.
              </p>
            </>
          )}

          {/* ==================================================
              SUCCESS
          ================================================== */}

          {status === "success" && (
            <>
              <div
                style={{
                  width: "56px",
                  height: "56px",
                  borderRadius: "16px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  background: "var(--success-soft)",
                  color: "var(--success)",
                  marginBottom: "20px",
                }}
              >
                <CheckCircle2 size={30} />
              </div>

              <h2 className="sh-auth-heading">
                Email verified
              </h2>

              <p className="sh-auth-subtitle">
                {message}
              </p>

              <Link
                href="/login"
                className="sh-btn sh-btn-primary sh-btn-full sh-btn-lg"
                style={{
                  marginTop: "24px",
                  textDecoration: "none",
                }}
              >
                Continue to login
              </Link>
            </>
          )}

          {/* ==================================================
              ERROR
          ================================================== */}

          {status === "error" && (
            <>
              <div
                style={{
                  width: "56px",
                  height: "56px",
                  borderRadius: "16px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  background: "var(--danger-soft)",
                  color: "var(--danger)",
                  marginBottom: "20px",
                }}
              >
                <XCircle size={30} />
              </div>

              <h2 className="sh-auth-heading">
                Verification failed
              </h2>

              <p className="sh-auth-subtitle">
                {message}
              </p>

              <div
                style={{
                  display: "flex",
                  gap: "10px",
                  marginTop: "24px",
                }}
              >
                <Link
                  href="/login"
                  className="sh-btn sh-btn-secondary sh-btn-full"
                  style={{
                    textDecoration: "none",
                  }}
                >
                  Back to login
                </Link>

                <Link
                  href="/resend-verification"
                  className="sh-btn sh-btn-primary sh-btn-full"
                  style={{
                    textDecoration: "none",
                  }}
                >
                  Resend email
                </Link>
              </div>
            </>
          )}

          {/* ==================================================
              WAITING
          ================================================== */}

          {status === "waiting" && (
            <>
              <div
                style={{
                  width: "56px",
                  height: "56px",
                  borderRadius: "16px",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  background: "var(--primary-soft)",
                  color: "var(--primary)",
                  marginBottom: "20px",
                }}
              >
                <Mail size={28} />
              </div>

              <h2 className="sh-auth-heading">
                Check your email
              </h2>

              <p className="sh-auth-subtitle">
                We sent a verification link to your
                email address. Open that link to
                activate your account.
              </p>

              <Link
                href="/resend-verification"
                className="sh-btn sh-btn-primary sh-btn-full sh-btn-lg"
                style={{
                  marginTop: "24px",
                  textDecoration: "none",
                }}
              >
                Resend verification email
              </Link>

              <p
                style={{
                  marginTop: "20px",
                  textAlign: "center",
                  color: "var(--text-muted)",
                  fontSize: "13px",
                }}
              >
                Already verified?{" "}

                <Link
                  href="/login"
                  className="sh-link"
                >
                  Sign in
                </Link>
              </p>
            </>
          )}
        </div>
      </section>
    </main>
  );
}

