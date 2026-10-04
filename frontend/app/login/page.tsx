"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import {
  Eye,
  EyeOff,
  Loader2,
  LockKeyhole,
  Mail,
} from "lucide-react";

import { apiRequest } from "@/lib/api";

type LoginResponse = {
  message?: string;
  accessToken?: string;
  refreshToken?: string;
  user?: {
    id: string;
    username: string;
    email: string;
    role: "USER" | "ANALYST" | "ADMIN";
  };
};

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const [showPassword, setShowPassword] = useState(false);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(
    event: FormEvent<HTMLFormElement>
  ) {
    event.preventDefault();

    setError("");
    setLoading(true);

    const identifier = email.trim();

    try {
      const response = await apiRequest<LoginResponse>(
        "/api/v1/auth/login",
        {
          method: "POST",
          body: JSON.stringify({
            identifier,
            password,
          }),
        }
      );

      /*
       * Temporary authentication storage.
       *
       * We will replace this with the final
       * HttpOnly-cookie/refresh-token flow before
       * production deployment.
       */

      if (response.accessToken) {
        sessionStorage.setItem(
          "accessToken",
          response.accessToken
        );
      }

      if (response.refreshToken) {
        sessionStorage.setItem(
          "refreshToken",
          response.refreshToken
        );
      }

      /*
       * Role-based frontend redirect.
       *
       * Real authorization is still performed
       * by the Go/Gin backend.
       */

      if (response.user?.role === "ADMIN") {
        window.location.href = "/admin";
      } else {
        window.location.href = "/dashboard";
      }
    } catch (err) {
      const message =
        err instanceof Error
          ? err.message
          : "Unable to sign in. Please try again.";

      /*
       * Unverified email
       *
       * Backend returns:
       *
       * 403
       * {
       *   "error": "email address is not verified"
       * }
       *
       * Redirect the user to the verification page
       * and carry the entered email address with the URL.
       */

      if (
        message
          .toLowerCase()
          .includes("email address is not verified")
      ) {
        window.location.href =
          `/verify-email?email=${encodeURIComponent(
            identifier
          )}`;

        return;
      }

      /*
       * Also support the shorter message in case
       * the backend message is changed later.
       */

      if (
        message
          .toLowerCase()
          .includes("email not verified")
      ) {
        window.location.href =
          `/verify-email?email=${encodeURIComponent(
            identifier
          )}`;

        return;
      }

      /*
       * All other login errors are displayed normally.
       */

      setError(message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="sh-auth-page">
      {/* =====================================================
          LEFT BRAND SECTION
          ===================================================== */}

      <section className="sh-auth-brand">
        <div className="sh-auth-brand-content">
          <span className="sh-auth-brand-badge">
            Social media management platform
          </span>

          <h1 className="sh-auth-brand-title">
            Your social media.
            <br />
            One workspace.
          </h1>

          <p className="sh-auth-brand-description">
            Create, publish, schedule, monitor, and
            analyze your social media activity from one
            centralized workspace.
          </p>

          <div className="sh-auth-brand-points">
            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                ✓
              </span>

              Manage multiple social accounts
            </div>

            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                ✓
              </span>

              Schedule and publish content
            </div>

            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                ✓
              </span>

              Monitor engagement and analytics
            </div>
          </div>
        </div>
      </section>

      {/* =====================================================
          RIGHT LOGIN SECTION
          ===================================================== */}

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

          {/* Heading */}

          <h2 className="sh-auth-heading">
            Welcome back
          </h2>

          <p className="sh-auth-subtitle">
            Sign in to continue to your workspace.
          </p>

          {/* Error */}

          {error && (
            <div
              className="sh-alert sh-alert-error"
              role="alert"
            >
              {error}
            </div>
          )}

          {/* Form */}

          <form
            onSubmit={handleSubmit}
            className="sh-form"
            style={{
              marginTop: error ? "18px" : "0",
            }}
          >
            {/* Email */}

            <div className="sh-field">
              <label
                htmlFor="email"
                className="sh-label"
              >
                Email address
              </label>

              <div className="sh-input-wrap">
                <Mail className="sh-input-icon" />

                <input
                  id="email"
                  name="email"
                  type="email"
                  value={email}
                  onChange={(event) =>
                    setEmail(event.target.value)
                  }
                  placeholder="you@example.com"
                  autoComplete="email"
                  required
                  disabled={loading}
                  className="sh-input has-icon"
                />
              </div>
            </div>

            {/* Password */}

            <div className="sh-field">
              <div className="sh-form-footer">
                <label
                  htmlFor="password"
                  className="sh-label"
                >
                  Password
                </label>

                <Link
                  href="/forgot-password"
                  className="sh-link"
                  style={{ fontSize: "12px" }}
                >
                  Forgot password?
                </Link>
              </div>

              <div className="sh-input-wrap">
                <LockKeyhole className="sh-input-icon" />

                <input
                  id="password"
                  name="password"
                  type={
                    showPassword
                      ? "text"
                      : "password"
                  }
                  value={password}
                  onChange={(event) =>
                    setPassword(event.target.value)
                  }
                  placeholder="Enter your password"
                  autoComplete="current-password"
                  required
                  disabled={loading}
                  className="sh-input has-icon has-action"
                />

                <button
                  type="button"
                  className="sh-input-action"
                  onClick={() =>
                    setShowPassword(
                      (current) => !current
                    )
                  }
                  aria-label={
                    showPassword
                      ? "Hide password"
                      : "Show password"
                  }
                  disabled={loading}
                >
                  {showPassword ? (
                    <EyeOff size={17} />
                  ) : (
                    <Eye size={17} />
                  )}
                </button>
              </div>
            </div>

            {/* Submit */}

            <button
              type="submit"
              disabled={loading}
              className="sh-btn sh-btn-primary sh-btn-full sh-btn-lg"
            >
              {loading ? (
                <>
                  <Loader2
                    size={17}
                    className="sh-spinner"
                  />

                  Signing in...
                </>
              ) : (
                "Sign in"
              )}
            </button>
          </form>

          {/* Register */}

          <p
            style={{
              marginTop: "24px",
              textAlign: "center",
              color: "var(--text-muted)",
              fontSize: "13px",
            }}
          >
            Don&apos;t have an account?{" "}
            <Link
              href="/register"
              className="sh-link"
            >
              Create an account
            </Link>
          </p>
        </div>
      </section>
    </main>
  );
}