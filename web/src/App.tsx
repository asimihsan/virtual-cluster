import React from 'react';
import { EventProvider } from './utils/EventContext';
import EventList from './components/EventList';

const App: React.FC = () => {
  return (
      // @ts-ignore
      <EventProvider>
        <div className="App">
          <EventList />
        </div>
      </EventProvider>
  );
};

export default App;
