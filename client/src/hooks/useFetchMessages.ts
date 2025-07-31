import { useState, useEffect } from "react";

export const useFetchMessages = () => {
    const [message, setMessage] = useState<string>('');
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchMessage = async () => {
            try {
                setLoading(true);
                setError(null);
                
                const response = await fetch("http://localhost:8080/message");
                
                if (!response.ok) {
                    throw new Error(`HTTP error! Status: ${response.status}`);
                }
                
                const data = await response.json();
                setMessage(data.message);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error');
            } finally {
                setLoading(false);
            }
        };

        fetchMessage();
    }, []);

    return { message, loading, error };
};
