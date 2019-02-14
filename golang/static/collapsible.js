let showPages = true;
let showCategories = true;

function tree() {
	let selected;
	const totalWidth = 800;
	const totalHeight = 800;
	const horizontalSpacing = 270;
	const circleRadius = 9;
	const circleGrowRadius = 14;
	var data,
		i = 0,
        duration = 750,
        margin = {top: 20, right: 0, bottom: 20, left: 80},
	    width = totalWidth - margin.left - margin.right,
        height = totalHeight - margin.top - margin.bottom,
        update;

	// const diagonal = d3.linkHorizontal().x(d => d.y).y(d => d.x);

	function updateChildren(d) {
		if (d.children) {
			d._children = d.children;
			d.children = null;
			return Promise.resolve(d);
		} else if (d._children) {
			d.children = d._children;
			d._children = null;
			return Promise.resolve(d);
		}

		if (!d.children && !d._children) {
			const name = d.data.name;
			const subcategoriesPromise = showCategories ? getSubcategoriesPromise(name) : Promise.resolve([]);
			return subcategoriesPromise
				.then((data) => {
					const hierarch = data.map(c => {
						const parent = d;
						const name = c.title.replace(/^Category:/, '');
						const children = null;
						const obj = { name, children };
						const n = d3.hierarchy(obj, x => x.children);
						n.parent = parent;
						n.depth = parent.depth + 1;
						return n;
					});
					d.children = hierarch.length > 0 ? hierarch : null;
					d.data.children = hierarch.length > 0 ? hierarch.map(d => d.data) : null;

					const pagePromise = showPages ? getPages(name) : Promise.resolve([]);
					return Promise.all([d, pagePromise]);
				})
				.then(([d, pages]) => {
					const hierarch = pages.map(c => {
						const parent = d;
						const name = c.title.replace(/^Category:/, '');
						const children = null;
						const obj = { name, type: 'page', children };
						const n = d3.hierarchy(obj, x => x.children);
						n.parent = parent;
						n.depth = parent.depth + 1;
						return n;
					});
					d.children = hierarch.length > 0 ? d.children.concat(hierarch) : d.children;
					d.data.children = hierarch.length > 0 ? d.data.children.concat(hierarch.map(d => d.data)) : d.data.children;
					return d;
				});
		}

		return Promise.resolve(d);
	}


    function chart(selection){
        selection.each(function () {
        	height = height - margin.top - margin.bottom;
			width = width - margin.left - margin.right;

			const zoomListener = d3.zoom()
				.scaleExtent([1 / 2, 4])
				.on("zoom", zoomed);

            // append the svg object to the selection
			const svg = d3.select(this)
				.append('svg')
                .attr('width', '100%') //width + margin.left + margin.right)
				.attr('height', '100%') //height + margin.top + margin.bottom)
		    	.attr('preserveAspectRatio', 'xMinYMin meet')
				.style("user-select", "none")
				.call(zoomListener);

			const g = svg.append('g');
			g.attr('transform', `translate(${margin.left}, ${margin.top})`);

			function zoomed() {
				g.attr("transform", d3.event.transform);
			}

            // declares a tree layout and assigns the size of the tree
			const treemap = d3.tree().size([height, width]);

			// assign parent, children, height, depth
			var root = d3.hierarchy(data, d => d.children);
			root.x0 = height / 2; // left edge of the rectangle
			root.y0 = 0; // top edge of the triangle

			function transform() {
				console.log('transform', this);
				return d3.zoomIdentity
					.translate(selected.x0 - horizontalSpacing, selected.y0);
			}

			// toggle children on click
			function click(d) {
				const node = d3.select(this).select('circle');

				function blink() {
					node.transition()
							.ease(d3.easeCubic)
							.duration(300)
							.style('fill', '#F2BF00')
							.style('stroke', '#D99100')
						.transition()
							.ease(d3.easeCubic)
							.duration(300)
							.style('fill', '#888')
							.style('stroke', '#888')
						// .transition()
						//	.ease(d3.easeCubic)
						//	.duration(300)
						//	.style('fill', '#ffa')
						.on('end', blink);
				}

				blink();
				updateRemote(d).finally(() => {
					console.log(node);
					node.transition()
							.ease(d3.easeCubic)
							.duration(300)
							.style('fill', '#C7FF0B')
							.style('stroke', '#004B19')
							.attr('r', circleGrowRadius)
						.transition()
							.delay(200)
							.ease(d3.easeCubic)
							.duration(200)
							.attr('r', circleRadius);
				});
			}

			function updateRemote(d) {
				return updateChildren(d)
					.then((d) => {
						selected = d;
						update(d);
						/*
						svg.transition()
							.delay(100)
							.duration(1000)
							.call(zoomListener.transform, transform)
						*/
					});
			}

	        function update(source) {
				// assigns the x and y position for the nodes
				var treeData = treemap(root);
				console.log(treeData);

		        // compute the new tree layout
		        var nodes = treeData.descendants(),
		            links = treeData.descendants().slice(1);

		        // normalise for fixed depth
		        nodes.forEach(d => d.y = d.depth * horizontalSpacing);

		        // ****************** Nodes section ***************************

		        // update the nodes ...
		        var node = svg.selectAll('g.node')
			        .data(nodes, d => d.id || (d.id = ++i));

		        // Enter any new modes at the parent's previous position.
		        var nodeEnter = node.enter().append('g')
					.attr('class', 'node')
			        .attr('transform', function(d) {
						const x = source.y0 + margin.top;
						const y = source.x0 + margin.left;
			        	return `translate(${x}, ${y})`;
			        })
			        .on('click', click);

		        // add circle for the nodes
		        nodeEnter.append('circle')
					.attr('class', 'node')
			        .attr('r', 1e-6);

		        // add labels for the nodes
				nodeEnter.append('text')
					.attr('class', 'label')
			        .attr('dy', '.35em')
			        .attr('x', d =>
			        	d.children || d._children ? 0 : 13
			        )
			        .attr('y', d =>
			        	d.children || d._children ? -margin.top : 0
			        )
					.attr('text-anchor', d =>
						d.children || d._children ? 'middle' : 'start'
			        )
			        .text(d => d.data.name);

		        // add number of children to node circle
		        nodeEnter.append('text')
					.attr('class', 'node__label')
					.attr('x', -5)
			        .attr('y', 4)
			        .style('font-size', '10px');

		        // UPDATE
				var nodeUpdate = nodeEnter.merge(node);

				nodeUpdate.select('text.node__label')
					.text((d) => {
						if (d.children) {
							return d.children.length;
						}
						else if (d._children) {
							return d._children.length;
						}
					});

		        // transition to the proper position for the node
		        nodeUpdate.transition().duration(duration)
			        .attr('transform', (d) => {
						const x = d.y + margin.top;
						const y = d.x + margin.left;
			        	return `translate(${x}, ${y})`;
			        });

		        // update the node attributes and style
		        nodeUpdate.select('circle.node')
			        .attr('r', circleRadius)
					.attr('class', (d) => {
						if (d.children) { return 'node--internal'; };
						if (d.data.type === 'page') { return 'node--page'; };
						return 'node--leaf';
					})
			        .attr('cursor', 'pointer');

		        // remove any exiting nodes
		        var nodeExit = node.exit()
			        .transition().duration(duration)
			        .attr('transform', (d) => {
						const x = source.y + margin.top;
						const y = source.x + margin.left;
						return `translate(${x},${y})`;
			        })
			        .remove();

		        // on exit reduce the node circles size to 0
		        nodeExit.select('circle')
			        .attr('r', 1e-6);

		        // on exit reduce the opacity of text labels
		        nodeExit.select('text')
			        .style('fill-opacity', 1e-6);

		        // ****************** links section ***************************

		        // update the links
		        var link = svg.selectAll('path.link')
			        .data(links, d => d.id);

		        // enter any new links at the parent's previous position
		        var linkEnter = link.enter().insert('path', 'g')
			        .attr('class', 'link')
			        .attr('d', (d) => {
			        	var o = {x: source.x0 /* + margin.left */, y: source.y0 /* + margin.top */};
			        	return diagonal({ source: o, target: o });
			        });

		        // UPDATE
		        var linkUpdate = linkEnter.merge(link);

				// transition back to the parent element position
				// d => diagonal({ source: d, target: d.parent })
		        linkUpdate.transition().duration(duration)
					.attr('d', d => diagonal({ source: d, target: d.parent }));

		        // remove any exiting links
		        var linkExit = link.exit()
			        .transition().duration(duration)
			        .attr('d', (d) => {
			        	var o = {x: source.x, y: source.y};
			        	return diagonal({ source: o, target: o });
			        })
			        .remove();

		        // store the old positions for transition
		        nodes.forEach(function(d) {
		        	d.x0 = d.x; // + margin.left;
		        	d.y0 = d.y; // + margin.top;
				});


				// creates a curved (diagonal) path from parent to the child nodes
		        function diagonal(point) {
					const s = point.source;
					const d = point.target;
		        	path = 'M ' + (s.y + margin.top) + ' ' + (s.x + margin.left) +
					        'C ' + ((s.y + d.y + (margin.top * 2)) / 2) + ' ' + (s.x + margin.left) +
					        ', ' + ((s.y + d.y + (margin.top * 2)) / 2) + ' ' + (d.x + margin.left) +
					        ', ' + (d.y + margin.top) + ' ' + (d.x + margin.left);
		        	return path;
				}
			}

			function updateData() {
				var root = d3.hierarchy(data, d => d.children);
				root.x0 = height / 2; // left edge of the rectangle
				root.y0 = 0; // top edge of the triangle
			}

	        update(root, data);
        });
	}

	chart.data = function(value) {
    	if (!arguments.length) return data;
    	data = value;
    	if (typeof updateData === 'function') updateData();
    	return chart;
	};

	chart.width = function(value) {
    	if (!arguments.length) return width;
    	width = value;
    	if (typeof updateWidth === 'function') updateWidth();
    	return chart;
	};

    return chart;
}

function getSubcategoriesPromise(category) {
    const defaultCategory = 'Roman usurpers';
    const cat = category || defaultCategory;
    return d3.json(`/category/v1/${cat}`);
}

window.addEventListener('load', () => {
    const categoryInput = document.getElementById('category');
    const defaultCategory = 'Roman emperors';
	categoryInput.value = defaultCategory;

	document.getElementById('page-toggle').checked = true;
	document.getElementById('category-toggle').checked = true;

	// cyGraph(defaultCategory);

	const category = defaultCategory;
	const children = null
	const root = { name: category, children };
	const chart = tree().data(root);
	d3.select('#d3').call(chart);
});

function refreshChart(category) {
	const svg = d3.select("svg");
	svg.selectAll('*').remove();
	const children = null
	const root = { name: category, children };
	const chart = tree().data(root);
	svg.call(chart);
}

function readCategory() {
	const category = document.getElementById('category').value;
	return category;
}

const addButton = document.getElementById('add-category');
addButton.addEventListener('click', (event) => {
	const category = readCategory();
	refreshChart(category);
});

const pageToggle = document.getElementById('page-toggle');
pageToggle.addEventListener('click', (event) => {
	const checked = pageToggle.checked;
	const category = readCategory();
	showPages = checked;
	refreshChart(category)
});