import { useState, useEffect } from "react";
import { useAuth } from "../contexts/AuthContext"; // AuthContextからuseAuthをインポート

const useFetchMessages = () => {
    const [message, setMessage] = useState<string | null>(null);
    const { token } = useAuth(); // useAuthフックを使用してトークンを取得

    useEffect(() => {
        // トークンがない場合はAPIリクエストを行わない
        if (!token) {
            return;
        }

        fetch("http://localhost:8080/message", {
            method: "GET",
            mode: "cors",
            headers: {
                "Content-Type": "application/json",
                // Authorizationヘッダーにトークンを追加
                "Authorization": `Bearer ${token}`,
            },
        })
            .then((res) => {
                if (!res.ok) {
                    throw new Error(`HTTP error! Status: ${res.status}`);
                }
                return res.json();
            })
            .then((data) => {
                setMessage(data.message);
            })
            .catch((e) => console.error("Fetch error:", e));
    }, [token]); // tokenが変更されたときにuseEffectを再実行

    return message;
};

export default useFetchMessages;
