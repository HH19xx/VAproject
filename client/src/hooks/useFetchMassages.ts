import { useState, useEffect } from "react";

const useFetchMessages = () => {
    const [message, setMessage] = useState<string | null>(null);

    useEffect(() => {
        fetch("http://localhost:8080/message", {
            method: "GET",
            mode: "cors",
            headers: {
                "Content-Type": "application/json",
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
    }, []);

    return message;
};

export default useFetchMessages;
