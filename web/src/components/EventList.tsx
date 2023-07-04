import React from 'react';
import { useEventContext } from '../utils/EventContext';
import useEventService from '../services/useEventService';

const EventList: React.FC = () => {
    const { events } = useEventContext();

    useEventService();

    return (
        <div>
            {events.map((event, index) => (
                <div key={index}>
                    <p>{event.timestamp}</p>
                    <p>{event.type}</p>
                    {/* Add more fields as needed */}
                </div>
            ))}
        </div>
    );
};

export default EventList;
