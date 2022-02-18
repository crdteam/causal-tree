
export class Crdt {
    constructor(states) {
        this.states = states;
        this.time = 0;

        this.keyListener = this.handleKeypress.bind(this)
        window.addEventListener("keydown", this.keyListener)
    }
    
    destroy() {
        window.removeEventListener("keydown", this.keyListener)
    }

    handleKeypress(evt) {
        if (evt.defaultPrevented) {
            return
        }

        let handled = false
        switch(evt.key) {
        case "ArrowLeft":
            this.addTime(-1)
            handled = true
            break
        case "ArrowRight":
            this.addTime(+1)
            handled = true
            break
        }

        if (handled) {
            evt.preventDefault()
        }
    }

    render() {
        let state = this.state()
        $("#crdt")
            .html("")
            .append(this.controls())
            .append($("<div>")
                .addClass("state")
                .append($("<h2>")
                    .append(state['Action']))
                .append(this.renderSites(state['Sites'])))
    }

    renderSites(sites) {
        let siteEls = []
        let i = 0
        for (let site of sites) {
            siteEls.push(this.renderSite(site, i))
            i++
        }
        return siteEls
    }

    renderSite(site, i) {
        let index = site['Sitemap'].indexOf(site['SiteID']);
        return $("<div>")
            .addClass("site")
            .append($("<h3>").append(`List #${i}`))
            .append($("<h4>").append("Sitemap"))
            .append($("<ol>")
                .addClass("sitemap")
                .attr("start", 0)
                .append(this.renderSiteIDs(site['Sitemap'], index)))
            .append($("<h4>").append("Weave"))
            .append($("<div>")
                .addClass("weave")
                .append(this.renderAtoms(site['Weave'])))
    }

    renderSiteIDs(sitemap, index) {
        let ids = []
        let i = 0
        for (let siteID of sitemap) {
            let el = $("<li>").append(siteID)
            if (i == index) {
                el.addClass("highlight")
            }
            ids.push(el)
            i++
        }
        return ids
    }

    renderAtoms(weave) {
        let atoms = []
        for (let atom of weave) {
            atoms.push(this.renderAtom(atom))
        }
        return atoms
    }

    renderAtom(atom) {
        return $("<div>")
            .addClass("atom")
            .append($("<div>")
                .addClass("atom-id")
                .text(this.idString(atom['ID'])))
            .append($("<div>")
                .addClass("atom-value")
                .text(atom['Value']))
            .append($("<div>")
                .addClass("atom-cause")
                .text(this.idString(atom['Cause'])))
    }

    controls() {
        return $("<div>")
            .addClass("controls")
            .append($("<button>").text("-100").click(() => this.addTime(-100)))
            .append($("<button>").text("-10").click(() => this.addTime(-10)))
            .append($("<button>").text("-1").click(() => this.addTime(-1)))
            .append($("<input>")
                .attr("type", "number")
                .val(this.time)
                .change((evt) => this.goToTime(evt.target.valueAsNumber)))
            .append($("<button>").text("+1").click(() => this.addTime(+1)))
            .append($("<button>").text("+10").click(() => this.addTime(+10)))
            .append($("<button>").text("+100").click(() => this.addTime(+100)))
    }

    addTime(dt) {
        this.goToTime(this.time+dt)
    }

    goToTime(t) {
        if (t < 0) {
            t = 0
        }
        if (t > this.states.length-1) {
            t = this.states.length-1
        }
        if (t != this.time) {
            this.time = t
            this.render()
        }
    }

    state(time) {
        if (time === undefined) {
            time = this.time
        }
        return this.states[time]
    }

    idString(id) {
        return `S${id['Site']}@T${id['Timestamp']}`
    }
}

