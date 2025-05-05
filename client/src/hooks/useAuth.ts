import { useState, useEffect } from "react";

const useAuth = () => {
  const [token, setToken] = useState<string | null>(null);
  const [userID, setUserID] = useState<number | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  // 初期化（localStorageからトークン読み込み）
  useEffect(() => {
    try {
      const storedToken = window.localStorage.getItem("token");
      if (storedToken) setToken(storedToken);
    } catch (e) {
      console.warn("localStorageへのアクセスが制限されています", e);
    }
  }, []);

  // トークンがあるなら /me に問い合わせ
  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }

    fetch("http://localhost:8080/me", {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${token}`,
      }
    })
      .then(res => {
        if (!res.ok) throw new Error("未認証");
        return res.json();
      })
      .then(data => {
        setUserID(data.user_id);
      })
      .catch(() => {
        window.localStorage.removeItem("token");
        setToken(null);
        setUserID(null);
      })
      .finally(() => setLoading(false));
  }, [token]);

  const login = async (name: string, password: string) => {
    const res = await fetch("http://localhost:8080/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, password })
    });

    if (!res.ok) throw new Error("ログイン失敗");

    const data = await res.json();
    window.localStorage.setItem("token", data.token);
    setToken(data.token);
  };

  const logout = () => {
    window.localStorage.removeItem("token");
    setToken(null);
    setUserID(null);
  };

  return { token, userID, login, logout, loading };
};

export default useAuth;
