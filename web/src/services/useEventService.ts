import { useEffect } from 'react';
import { LogEvent, HttpRequestEvent, KafkaMessageEvent } from '../models/Event';
import { useEventContext } from '../utils/EventContext';

const useEventService = () => {
    const { addEvent } = useEventContext();

    useEffect(() => {
        const websocket = new WebSocket('ws://localhost:1371/ws');

        websocket.onmessage = (event) => {
            if (event.data instanceof Blob) {
                const reader = new FileReader();
                reader.onload = function() {
                    const data = JSON.parse(<string>this.result);
                    console.log(data);

                    switch (data.type) {
                        case 'log':
                            addEvent(data as LogEvent);
                            break;
                        case 'http_request':
                            addEvent(data as HttpRequestEvent);
                            break;
                        case 'kafka_message':
                            addEvent(data as KafkaMessageEvent);
                            break;
                        default:
                            console.error(`Unknown event type: ${data.type}`);
                    }
                };
                reader.readAsText(event.data);
            } else {
                console.error('Received data is not a Blob');
            }
        };

        return () => {
            if (websocket.readyState === 1) {
                websocket.close();
            }
        };
    }, [addEvent]);

    // Other service methods can go here
};

export default useEventService;
