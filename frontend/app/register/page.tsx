"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import {
  CheckCircle2,
  Eye,
  EyeOff,
  Loader2,
  User,
  Mail,
  LockKeyhole,
} from "lucide-react";

import { apiRequest } from "@/lib/api";

type RegisterResponse = {
  message?: string;
};

export default function RegisterPage() {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] =
    useState("");

  const [showPassword, setShowPassword] =
    useState(false);

  const [showConfirmPassword, setShowConfirmPassword] =
    useState(false);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(
    event: FormEvent<HTMLFormElement>
  ) {
    event.preventDefault();

    setError("");

    /*
     * Client-side password confirmation.
     * This is only for user experience.
     * The backend remains responsible for
     * validating the actual password.
     */
    if (password !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    setLoading(true);

    try {
      /*
       * IMPORTANT:
       * We intentionally do NOT send:
       *
       * role
       * status
       * emailVerified
       *
       * These are backend-controlled fields.
       */
      await apiRequest<RegisterResponse>(
        "/api/v1/auth/register",
        {
          method: "POST",
          body: JSON.stringify({
            firstName,
            lastName,
            username,
            email,
            password,
          }),
        }
      );

      /*
       * Registration succeeded.
       *
       * Send the user to the verification page.
       * The email is only used by the frontend to
       * display which email address should be checked.
       */
      const encodedEmail =
        encodeURIComponent(email);

      window.location.href =
        `/verify-email?email=${encodedEmail}`;
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to create your account."
      );

      setLoading(false);
    }
  }

  return (
    <main className="sh-auth-page">
      {/* ======================================================
          BRAND PANEL
      ====================================================== */}

      <section className="sh-auth-brand">
        <div className="sh-auth-brand-content">
          <span className="sh-auth-brand-badge">
            Welcome to SocialHub
          </span>

          <h1 className="sh-auth-brand-title">
            Everything social.
            <br />
            In one place.
          </h1>

          <p className="sh-auth-brand-description">
            Connect your social accounts, organize
            content, schedule publications, and monitor
            your audience from one clean workspace.
          </p>

          <div className="sh-auth-brand-points">
            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                <CheckCircle2 size={16} />
              </span>

              Centralized social account management
            </div>

            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                <CheckCircle2 size={16} />
              </span>

              Content publishing and scheduling
            </div>

            <div className="sh-auth-brand-point">
              <span className="sh-auth-brand-point-icon">
                <CheckCircle2 size={16} />
              </span>

              Engagement and performance monitoring
            </div>
          </div>
        </div>
      </section>

      {/* ======================================================
          REGISTER PANEL
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

          {/* Heading */}

          <h2 className="sh-auth-heading">
            Create your account
          </h2>

          <p className="sh-auth-subtitle">
            Start managing your social media from one
            workspace.
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
              marginTop: error ? 18 : 0,
            }}
          >
            {/* ==================================================
                FIRST NAME / LAST NAME
            ================================================== */}

            <div className="sh-form-row">
              {/* First name */}

              <div className="sh-field">
                <label
                  htmlFor="firstName"
                  className="sh-label"
                >
                  First name
                </label>

                <div className="sh-input-wrap">
                  <User className="sh-input-icon" />

                  <input
                    id="firstName"
                    name="firstName"
                    type="text"
                    value={firstName}
                    onChange={(event) =>
                      setFirstName(
                        event.target.value
                      )
                    }
                    placeholder="First name"
                    autoComplete="given-name"
                    required
                    disabled={loading}
                    className="sh-input has-icon"
                  />
                </div>
              </div>

              {/* Last name */}

              <div className="sh-field">
                <label
                  htmlFor="lastName"
                  className="sh-label"
                >
                  Last name
                </label>

                <input
                  id="lastName"
                  name="lastName"
                  type="text"
                  value={lastName}
                  onChange={(event) =>
                    setLastName(
                      event.target.value
                    )
                  }
                  placeholder="Last name"
                  autoComplete="family-name"
                  required
                  disabled={loading}
                  className="sh-input"
                />
              </div>
            </div>

            {/* ==================================================
                USERNAME
            ================================================== */}

            <div className="sh-field">
              <label
                htmlFor="username"
                className="sh-label"
              >
                Username
              </label>

              <div className="sh-input-wrap">
                <User className="sh-input-icon" />

                <input
                  id="username"
                  name="username"
                  type="text"
                  value={username}
                  onChange={(event) =>
                    setUsername(
                      event.target.value
                    )
                  }
                  placeholder="Choose a username"
                  autoComplete="username"
                  required
                  disabled={loading}
                  className="sh-input has-icon"
                />
              </div>
            </div>

            {/* ==================================================
                EMAIL
            ================================================== */}

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
                    setEmail(
                      event.target.value
                    )
                  }
                  placeholder="you@example.com"
                  autoComplete="email"
                  required
                  disabled={loading}
                  className="sh-input has-icon"
                />
              </div>
            </div>

            {/* ==================================================
                PASSWORD
            ================================================== */}

            <div className="sh-field">
              <label
                htmlFor="password"
                className="sh-label"
              >
                Password
              </label>

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
                    setPassword(
                      event.target.value
                    )
                  }
                  placeholder="Create a password"
                  autoComplete="new-password"
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

            {/* ==================================================
                CONFIRM PASSWORD
            ================================================== */}

            <div className="sh-field">
              <label
                htmlFor="confirmPassword"
                className="sh-label"
              >
                Confirm password
              </label>

              <div className="sh-input-wrap">
                <LockKeyhole className="sh-input-icon" />

                <input
                  id="confirmPassword"
                  name="confirmPassword"
                  type={
                    showConfirmPassword
                      ? "text"
                      : "password"
                  }
                  value={confirmPassword}
                  onChange={(event) =>
                    setConfirmPassword(
                      event.target.value
                    )
                  }
                  placeholder="Confirm your password"
                  autoComplete="new-password"
                  required
                  disabled={loading}
                  className="sh-input has-icon has-action"
                />

                <button
                  type="button"
                  className="sh-input-action"
                  onClick={() =>
                    setShowConfirmPassword(
                      (current) => !current
                    )
                  }
                  aria-label={
                    showConfirmPassword
                      ? "Hide password"
                      : "Show password"
                  }
                  disabled={loading}
                >
                  {showConfirmPassword ? (
                    <EyeOff size={17} />
                  ) : (
                    <Eye size={17} />
                  )}
                </button>
              </div>
            </div>

            {/* ==================================================
                SUBMIT
            ================================================== */}

            <button
              type="submit"
              disabled={loading}
              className="sh-btn sh-btn-primary sh-btn-full sh-btn-lg"
            >
              {loading ? (
                <>
                  <Loader2
                    size={17}
                    className="animate-spin"
                  />

                  Creating account...
                </>
              ) : (
                "Create account"
              )}
            </button>
          </form>

          {/* Login link */}

          <p
            style={{
              marginTop: 24,
              textAlign: "center",
              color: "var(--text-muted)",
              fontSize: 13,
            }}
          >
            Already have an account?{" "}
            <Link
              href="/login"
              className="sh-link"
            >
              Sign in
            </Link>
          </p>
        </div>
      </section>
    </main>
  );
}