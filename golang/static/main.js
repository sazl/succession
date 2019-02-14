let cy;
let globalId = 0;
const layoutOptions = {
    name: 'klay',
};

function getJSON(url) {
    return fetch(url)
        .then(function(response) {
            return response.json();
        });
}

function getCategory(cmtitle) {
    return getJSON(`category/v1/${cmtitle}`);
}

function getSubCategories(cmtitle) {
    return getJSON(`category/${cmtitle}/subcategories`);
}

function getPages(cmtitle) {
    return getJSON(`category/${cmtitle}/pages`);
}

function makeSubCategory(cat, parentId) {
    const nodeId = globalId++;
    const category = cat.title.trim().replace(/Category:/g, '');

    const node = {
        group: 'nodes',
        data: {
            id: nodeId,
            label: category,
            type: 'category',
            category: category,
            // parent: parentId,
        }
    }
    const edge = {
        group: 'edges',
        data: {
            source: parentId,
            target: nodeId,
        },
    };

    return [node, edge];
}

function makePage(page, parentId) {
    const label = page.title.trim().replace(/Category:/g, '');

    const node = {
        group: 'nodes',
        data: {
            id: page.pageid,
            label: label,
            type: 'page',
            page: page.title,
            // parent: parentId,
        }
    }
    const edge = {
        group: 'edges',
        data: {
            source: parentId,
            target: page.pageid,
        },
    };

    return [node, edge];
}

function cyGraph(category) {
    Promise.all([getSubCategories(category), getPages(category)])
        .spread((subcats, pages) => {
            const parentId = globalId++;
            const parent = {
                group: 'nodes',
                data: {
                    id: parentId,
                    label: category,
                    type: 'category',
                }
            };

            const subcatNodes = subcats.flatMap((cat) => makeSubCategory(cat, parentId));
            const pageNodes = pages.flatMap((page) => makePage(page, parentId));
            const allNodes = [].concat([parent], subcatNodes, pageNodes);

            const els = cy.add(allNodes);
            const elsLayout = els.layout(layoutOptions);
            elsLayout.run();
        });
}

/*
document.addEventListener('DOMContentLoaded', function(){
    cy = cytoscape({
        container: document.getElementById('cy'),
        boxSelectionEnabled: false,
        autounselectify: true,
        layout: layoutOptions,
        style: [
            {
                selector: 'node[label]',
                style: {
                    label: 'data(label)',
                    'font-size': '8px',
                }
            },
            {
                selector: 'node[type="category"]',
                style: {
                    shape: 'diamond',
                }
            },
            {
                selector: 'node[type="page"]',
                style: {
                    shape: 'square',
                    'background-color': '#ccd',
                }
            },
            {
                selector: 'edge',
                style: {
                    'curve-style': 'bezier',
                    'target-arrow-shape': 'triangle',
                    'opacity': 0.8,
                    'width': 0.5,
                }
            }
        ]
    });

    cy.on('tap', 'node[type="category"]', (evt) => {
        const node = evt.target;
        const data = node.data();
        const parentId = data.id;
        const category = data.category;

        Promise.all([getSubCategories(category), getPages(category)])
            .spread((subcats, pages) => {
                const subcatNodes = subcats.flatMap((cat) => makeSubCategory(cat, parentId));
                const pageNodes = pages.flatMap((page) => makePage(page, parentId));
                const allNodes = [].concat(subcatNodes, pageNodes);
                const els = cy.add(allNodes);

                cy.elements().layout(layoutOptions).run();
                cy.fit(els);
            });
    });

});
*/