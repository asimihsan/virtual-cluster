import React from 'react';
import { LogEvent, HttpRequestEvent, KafkaMessageEvent } from '../models/Event';

type Event = LogEvent | HttpRequestEvent | KafkaMessageEvent;

interface EventContextValue {
    events: Event[];
    addEvent: (event: Event) => void;
}

const EventContext = React.createContext<EventContextValue | undefined>(undefined);

// @ts-ignore
export const EventProvider: React.FC = ({ children }) => {
    const [events, setEvents] = React.useState<Event[]>([]);

    const addEvent = (event: Event) => {
        setEvents((prevEvents) => [...prevEvents, event].sort((a, b) => a.timestamp.localeCompare(b.timestamp)));
    };

    return (
        <EventContext.Provider value={{ events, addEvent }}>
            {children}
        </EventContext.Provider>
    );
};

export const useEventContext = () => {
    const context = React.useContext(EventContext);
    if (!context) {
        throw new Error('useEventContext must be used within an EventProvider');
    }
    return context;
};
