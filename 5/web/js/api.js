const api = {
    base: window.location.origin,

    async getEvents() {
        const res = await fetch(`${this.base}/events`);
        return res.ok ? res.json() : [];
    },

    async getEvent(id) {
        const res = await fetch(`${this.base}/events/${id}`);
        return res.ok ? res.json() : null;
    },

    async createEvent(title, date, total_seats) {
        const res = await fetch(`${this.base}/events`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ title, date: new Date(date).toISOString(), total_seats }),
        });
        return res.ok ? res.json() : null;
    },

    async bookSeat(eventId, user_name) {
        const res = await fetch(`${this.base}/events/${eventId}/book`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ user_name }),
        });
        return res.ok ? res.json() : null;
    },

    async confirmBooking(eventId, booking_id) {
        const res = await fetch(`${this.base}/events/${eventId}/confirm`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ booking_id }),
        });
        return res.ok;
    },
};
