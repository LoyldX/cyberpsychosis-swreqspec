// รองรับ FR-AUTH-01, CON-TECH-02: สถานะการเข้าสู่ระบบของสมาชิกทั้งเว็บ (ไม่มี logout ตาม ASM-03)
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { clearToken, fetchSession, readToken, saveToken, type Member } from "./api";

export type AuthStatus = "loading" | "authenticated" | "unauthenticated" | "unavailable";

export type AuthContextValue = {
  status: AuthStatus;
  member: Member | null;
  /** เก็บ Refresh Token แล้วตรวจ session ใหม่ (ใช้หลังเข้าสู่ระบบสำเร็จ) */
  signIn: (refreshToken: string) => Promise<void>;
  /** ตรวจ session ใหม่ด้วย token เดิม (ใช้ตอน status = unavailable) */
  retry: () => Promise<void>;
};

type AuthState = { status: AuthStatus; member: Member | null };

const AuthContext = createContext<AuthContextValue | null>(null);

// รองรับ FR-AUTH-01, ASM-05, ASM-07: อ่าน token → ตรวจ session → ตั้งสถานะ (401 ลบ token, 503/เครือข่ายล้ม ไม่ลบ)
export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({ status: "loading", member: null });
  // กันผลของคำขอเก่าทับผลของคำขอใหม่ (เช่น signIn/retry ซ้อนกัน หรือ unmount)
  const requestId = useRef(0);

  const checkSession = useCallback(async () => {
    const id = ++requestId.current;
    const token = readToken();
    if (!token) {
      setState({ status: "unauthenticated", member: null });
      return;
    }
    setState({ status: "loading", member: null });
    const result = await fetchSession(token);
    if (id !== requestId.current) return;
    if (result.kind === "ok") {
      setState({ status: "authenticated", member: result.member });
    } else if (result.kind === "unauthorized") {
      clearToken();
      setState({ status: "unauthenticated", member: null });
    } else {
      setState({ status: "unavailable", member: null });
    }
  }, []);

  useEffect(() => {
    void checkSession();
    return () => {
      requestId.current++;
    };
  }, [checkSession]);

  const signIn = useCallback(
    async (refreshToken: string) => {
      saveToken(refreshToken);
      await checkSession();
    },
    [checkSession],
  );

  const value = useMemo<AuthContextValue>(
    () => ({ status: state.status, member: state.member, signIn, retry: checkSession }),
    [state, signIn, checkSession],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// รองรับ FR-AUTH-01: hook อ่านสถานะ/สั่ง signIn, retry (ต้องอยู่ใต้ AuthProvider)
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth ต้องใช้ภายใต้ AuthProvider");
  return ctx;
}
