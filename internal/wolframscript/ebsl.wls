#!/usr/bin/env wolframscript -charset UTF8 -print All

On[Assert]

(* List fold-right with a left-identity element. Written this way to optimize out the left-identity anywhere in the list. *)
FoldRLI[f_,id_,{}]:=id
FoldRLI[f_,id_,{x_}]:=x
FoldRLI[f_,id_,l_]:=FoldRLIcons[f,id,First[l],Rest[l]]
FoldRLIcons[f_,id_,id_,xs_]:=FoldRLI[f,id,xs]
FoldRLIcons[f_,id_,x_,xs_]:=f[x,FoldRLI[f,id,xs]]

(* List fold-right with an identity element (on both the left and right). Written this way to optimize out the identity anywhere in the list. *)
FoldRI[f_,id_,{}]:=id
FoldRI[f_,id_,{x_}]:=x
FoldRI[f_,id_,{id_,x_}]:=x
FoldRI[f_,id_,{x_,id_}]:=x
FoldRI[f_,id_,l_]:=FoldRIcons[f,id,First[l],Rest[l]]
FoldRIcons[f_,id_,id_,xs_]:=FoldRI[f,id,xs]
FoldRIcons[f_,id_,x_,xs_]:=f[x,FoldRI[f,id,xs]]

(* List fold-left with a left-identity element. Written this way to optimize out the left-identity wherever possible in the list. *)
FoldLLI[f_,id_,{}]:=id
FoldLLI[f_,id_,l_]:=FoldLLICons[f,id,First[l],Rest[l]]
FoldLLICons[f_,id_,x_,{}]:=x
FoldLLICons[f_,id_,id_,xs_]:=FoldLLICons[f,id,First[xs],Rest[xs]]
FoldLLICons[f_,id_,x_,xs_]:=FoldLLICons[f,id,f[x,First[xs]],Rest[xs]]

(* List fold-left with an identity element (on both the left and right). Written this way to optimize out the identity anywhere in the list. Also takes and optimizes out two absorbing elements. *)
FoldLI[f_,id_,a1_,a2_,{}]:=id
FoldLI[f_,id_,a1_,a2_,l_]:=FoldLICons[f,id,a1,a2,First[l],Rest[l]]
FoldLICons[f_,id_,a1_,a2_,x_,{}]:=x
FoldLICons[f_,id_,a1_,a2_,id_,xs_]:=FoldLICons[f,id,a1,a2,First[xs],Rest[xs]]
FoldLICons[f_,id_,a1_,a2_,a1_,xs_]:=a1
FoldLICons[f_,id_,a1_,a2_,a2_,xs_]:=a2
FoldLICons[f_,id_,a1_,a2_,x_,xs_]:=FoldLIConsCons[f,id,a1,a2,x,First[xs],Rest[xs]]
FoldLIConsCons[f_,id_,a1_,a2_,x_,id_,xs_]:=FoldLICons[f,id,a1,a2,x,xs]
FoldLIConsCons[f_,id_,a1_,a2_,x_,a1_,xs_]:=a1
FoldLIConsCons[f_,id_,a1_,a2_,x_,a2_,xs_]:=a2
FoldLIConsCons[f_,id_,a1_,a2_,x_,x2_,xs_]:=FoldLICons[f,id,a1,a2,f[x,x2],xs]

(* Given a list of distinct elements, generate all pairs of distinct elements with the order being insignificant (that is, the list of combinations of length 2 of elements from from the given list). *)
DistinctPairs[{}]:={}
DistinctPairs[{x_}]:={}
DistinctPairs[l_]:=Module[{r=Rest[l]},Join[Map[({First[l],#})&,r],DistinctPairs[r]]]

(* Determine whether the given list of lists is pairwise disjoint. *)
PairwiseDisjoint[l_]:=AllTrue[DistinctPairs[l],(Intersection[#[[1]],#[[2]]]=={})&]

(* Extract the key part of an association element. *)
AssociationElemKey[k_->v_]:=k

(* Extract the value part of an association element. *)
AssociationElemValue[k_->v_]:=v

(* A form of AssociationMap which passes the key and value to the provided function and returns an association with the same keys (the provided function only returns a new value, not a key-value pair. *)
AssociationMapFixed[f_,a_]:=AssociationMap[(#[[1]]->f[#[[1]],#[[2]]])&,a]


(* Find all the minimal-length lists in the given list of lists. *)
ShortestLists[l_]:=Module[{lengths,minlength,lists},
If[l=={},
{},
(lengths=Map[Length,l];
minlength=Min[lengths];
lists=Select[l,(Length[#]==minlength)&];
Assert[lists!={}];
lists)
]
]

(* This series of functions comprises a replacement for Mathematica's built-in FindVertexCut, which appears to be broken.  Some of the examples in this paper run across the following bug:
https://mathematica.stackexchange.com/questions/125146/findvertexcut-findedgecut
See also:
https://community.wolfram.com/groups/-/m/t/1327325
This is a tremendously inefficient algorithm; I presume that Mathematica's internal one is far more efficient, but it is apparently also buggy. Proper suitable algorithms for a real implementation might include Karger's algorithm or the Ford-Fulkerson algorithm.
The caller must guarantee that there is at least one path from v to w in g, and that there is no direct edge from v to w (so that there is some cut strictly between them).
*)
IsVertexCut[g_,v_,w_,vcs_]:=
(Assert[v!=w];
Assert[MemberQ[EdgeList[g],v\[DirectedEdge]w]==False];
Assert[FindPath[g,v,w]!={}];
FindPath[VertexDelete[g,vcs],v,w]=={})
FindVertexCutsOfLength[g_,v_,w_,n_]:=Module[{ivs,lengths,minlength,shorts},
Assert[n>=1];
Assert[n<=VertexCount[g]-2];
Assert[v!=w];
Assert[MemberQ[EdgeList[g],v\[DirectedEdge]w]==False];
Assert[FindPath[g,v,w]!={}];
ivs=Select[VertexList[g],((#!=v)\[And](#!=w))&];
Select[Subsets[ivs,n],((#!={})\[And]IsVertexCut[g,v,w,#])&]
]
FindAllVertexCuts[g_,v_,w_]:=Module[{maxCutLen=VertexCount[g]-2},
Assert[maxCutLen>=1];
Catenate[Map[(FindVertexCutsOfLength[g,v,w,#])&,Range[1..VertexCount[g]-2]]]
]
FindMinVertexCutsOfAtLeastLength[g_,v_,w_,n_]:=Module[{vcs},
vcs=FindVertexCutsOfLength[g,v,w,n];
If[vcs=={},(Assert[n<VertexCount[g]-2];FindMinVertexCutsOfAtLeastLength[g,v,w,n+1]),vcs]
]
FindMinVertexCuts[g_,v_,w_]:=FindMinVertexCutsOfAtLeastLength[g,v,w,1]
FindOneMinVertexCut[g_,v_,w_]:=Module[{vcs=FindMinVertexCuts[g,v,w]},
Assert[vcs!={}];
vcs[[1]]
]
Unprotect[FindVertexCut];Clear[FindVertexCut]
FindVertexCut[g_,v_,w_]:=FindOneMinVertexCut[g,v,w]

(* Is xs a subset (meaning, not necessarily respecting order, so not necessarily a sublist) of ys? *)
IsSubset[xs_,ys_]:=SubsetQ[ys,xs]

(* Are the two lists equal as sets? *)
SetsEqual[xs_,ys_]:=ContainsExactly[xs,ys]

(* Is xs a subset of any of the list of sets in l? *)
IsSubsetAny[xs_,l_]:=AnyTrue[l,(IsSubset[xs,#]==True)&]

(* Find all members of a list of sets which are subsets of any of the other members. *)
FindNonmaximal[l_]:=Select[l,(IsSubsetAny[#,Complement[l,{#}]])&]

(* Remove from a list of sets all the sets that are subsets of any others in the list. *)
DeleteNonmaximal[l_]:=With[{ld=DeleteDuplicates[l]},Complement[ld,FindNonmaximal[ld]]]

(* Find all the maximal subsets of some given subset which meet a given condition (that is, all the subsets which meet the condition and of which no strict superset meets the condition). *)
SelectMaximalSubsets[cond_,set_]:=DeleteNonmaximal[Select[Subsets[set],cond]]

(* Accessors for belief vectors. *)
belief[{xb_,xd_,xu_}]:=xb
disbelief[{xb_,xd_,xu_}]:=xd
uncertainty[{xb_,xd_,xu_}]:=xu

(* The belief constraint operator. *)
OverVector[v_List]:=And[Length[v]==3,belief[v]\[Element]Reals,disbelief[v]\[Element]Reals,uncertainty[v]\[Element]Reals,belief[v]>=0,belief[v]<1,disbelief[v]>=0,disbelief[v]<1,uncertainty[v]>0,uncertainty[v]<=1,belief[v]+disbelief[v]+uncertainty[v]==1]

(* Validate each of a list of belief vectors. *)
\[Checkmark][l_]:=Fold[And,True,Map[(\!\(\*OverscriptBox[\(#\), \(\[RightVector]\)]\))&,l]]

(* The belief vector representing complete uncertainty. *)
Uncertain:={0,0,1};

(* The belief vector representing certain belief. This is not actually valid as a belief vector, because it's not a belief -- it represents mathematical proof of a statement. *)
Proven:={1,0,0};

(* The belief vector representing complete disbelief. This is different from and weaker than saying that it is provably contradictory (saying that P is contradictory is expressed by saying that \[Not]P is Proven). Rather, it is the analogue in this intuitionistic logic with uncertainty of saying that a statement is unprovable in standard intuitionistic logic. A user who specifies this belief vector with respect to a proposition P is not saying "I am sure that \[Not]P is provable", but rather, "there is no evidence you can show me that would convince me that P is true". *)
Unsupportable:={0,1,0};

(* Define interpretations for certain symbols as the belief operators that we are going to define so that we may write symbolic expressions using those symbols without interpreting them, then evaluate them by interpreting them. *)
Unprotect[Cross];Clear[Cross]
EvalBelief[expr_]:=Block[{\[ScriptF]=Unsupportable,\[ScriptT]=Proven,\[ScriptU]=Uncertain,CircleDot=dot,CenterDot=scale,CirclePlus=consensus,SuperDagger=not,CircleMinus=contradot,CircleTimes=propagate,Vee=join,Cross=defaultDiscount},expr]


(* Evaluate a belief expression possibly containing belief operator symbols, under given assumptions. The "Full" version tries more transformations (so it can sometimes discover things that the non-Full version can't) but is slower. *)
SimplifyBelief[expr_,as_]:=Simplify[EvalBelief[expr],EvalBelief[as]]
FullSimplifyBelief[expr_,as_]:=FullSimplify[EvalBelief[expr],EvalBelief[as]]



(* Attempt to find conditions under which a given statement is true under given assumptions -- if this returns True then the given statement is always true under the given assumptions.  This can sometimes find proofs that even FullSimplifyBelief can't, but at the cost of being even slower. *)
ResolveBelief[vars_,assumptions_,conclusion_]:=Resolve[ForAll[vars,EvalBelief[assumptions],EvalBelief[conclusion]]]



(* Validate a belief vector possibly containing belief operator symbols, under given assumptions. The "Full" version tries more transformations (so it can sometimes discover things that the non-Full version can't) but is slower. *)
ValidateBelief[v_,as_]:=SimplifyBelief[\!\(\*OverscriptBox[\(v\), \(\[RightVector]\)]\),as]
FullValidateBelief[v_,as_]:=FullSimplifyBelief[\!\(\*OverscriptBox[\(v\), \(\[RightVector]\)]\),as]



(* Attempt to find conditions under which a given belief vector is valid under given assumptions -- if this returns True then the given vector is always valid under the given assumptions.  This can sometimes find proofs that even FullValidateBelief can't, but at the cost of being even slower. *)
ResolveValidBelief[vars_,assumptions_,vector_]:=ResolveBelief[vars,assumptions,\!\(\*OverscriptBox[\(vector\), \(\[RightVector]\)]\)]



(* The consensus operator from the paper. This can be interpreted as the belief vector that results from combining two completely independent, disjoint perspectives (expressed as belief vectors) on the same proposition. *)
consensus[{xb_,xd_,xu_},{yb_,yd_,yu_}]:={xu*yb+yu*xb,xu*yd+yu*xd,xu*yu}/(xu+yu-xu*yu)


(* Consensus of a list. Written this way to optimize out consensus with Uncertain (which is the identity element for \[CirclePlus]). We also optimize out Proven and Unsupportable (which are absorbing elements). *)
\[GothicCapitalC][l_List]:=FoldLI[(#1\[CirclePlus]#2)&,Uncertain,Proven,Unsupportable,l]


(* Some parameters for graph-plotting that I like the looks of. *)
DefaultGraphPlot[g_]:=GraphPlot[g,{VertexLabeling->True,EdgeLabeling->Automatic,DirectedEdges->True}]
DefaultGraphPlotList[gs_]:=Map[DefaultGraphPlot,gs]



(* Convert a vertex path to a string separated by direct-edge symbols. *)
PathToString[{}]:="(empty)"
PathToString[{v_}]:=ToString[v]
PathToString[l_]:=ToString[First[l]]<>"\[DirectedEdge]"<>PathToString[Rest[l]]

(* Get the source vertex of a directed edge. *)
SourceVertex[v_\[DirectedEdge]w_]:=v


(* Get the target vertex of a directed edge. *)
TargetVertex[v_\[DirectedEdge]w_]:=w


(* Tets for equality of edges. *)
EdgesEqual[e1_,e2_]:=(SourceVertex[e1]==SourceVertex[e2])\[And](TargetVertex[e1]==TargetVertex[e2])



(* Add a vertex with a given set of properties. *)
VertexAddWithProperties[g_,n_,p_]:=SetProperty[{VertexAdd[g,n],n},p]
VertexAddNextWithProperties[g_,p_]:=VertexAddWithProperties[g,VertexCount[g]+1,p]



(* Check whether a graph contains a vertex. *)
GraphHasVertex[g_,v_]:=MemberQ[VertexList[g],v]



(* Check whether a graph contains an edge. *)
GraphHasEdge[g_,e_]:=MemberQ[EdgeList[g],e]



(* Does a vertex have the given property? *)
VertexHasProperty[g_,p_,n_]:=MemberQ[PropertyList[{g,n}],p]



(* Find a property of a vertex. *)
VertexProperty[g_,p_,n_]:=PropertyValue[{g,n},p]



(* Obtain the type of a vertex. *)
VertexType[g_,n_]:=VertexProperty[g,"vType",n]



(* Check whether a vertex has the given type. *)
VertexHasType[g_,n_,t_]:=VertexType[g,n]==t



(* Assert that a vertex has the given type. *)
AssertVertexType[g_,n_,t_]:=Assert[VertexHasType[g,n,t]]



(* Add an edge with a given set of properties. *)
EdgeAddWithProperties[g_,e_,p_]:=SetProperty[{EdgeAdd[g,e],e},p]



(* Does an edge have the given property? *)
EdgeHasProperty[g_,p_,a_,b_]:=MemberQ[PropertyList[{g,a\[DirectedEdge]b}],p]



(* Find a property of the edge from a to b. *)
EdgeProperty[g_,p_,a_,b_]:=PropertyValue[{g,a\[DirectedEdge]b},p]



(* Obtain the type of an edge. *)
EdgeType[g_,a_,b_]:=EdgeProperty[g,"eType",a,b]



(* Check whether an edge has the given type. *)
EdgeHasType[g_,a_,b_,t_]:=EdgeType[g,a,b]==t



(* Assert that an edge has the given type. *)
AssertEdgeType[g_,a_,b_,t_]:=Assert[EdgeHasType[g,a,b,t]]



(* An empty graph. *)
g0:=Graph[{},{}]



(* Get the label of a graph vertex by index. *)
GetLabel[g_,n_]:=PropertyValue[{g,n},VertexLabels]
(* Get the list of labels of a list of vertices. *)
GetLabels[g_,v_]:=Map[(GetLabel[g,#])&,v]



(* Get the label of an edge by source and target vertex. *)
GetLabel[g_,a_,b_]:=PropertyValue[{g,a\[DirectedEdge]b},EdgeLabels]



(* Get the "name" property of the given vertex. *)
VertexName[g_,n_]:=PropertyValue[{g,n},"name"]



(* Add a vertex (resp. edge) with a default name and label. *)
VertexAddLabeled[g_,n_]:=With[{l=ToString[n]},VertexAddWithProperties[g,n,{"name"->l,VertexLabels->l}]]
EdgeAddLabeled[g_,e_]:=With[{l=ToString[SourceVertex[e]]<>"\[DirectedEdge]"<>ToString[TargetVertex[e]]},EdgeAddWithProperties[g,e,{"name"->l,EdgeLabels->l}]]



(* Get the labels of a path which is given in vertex list representation. *)
PathLabels[g_,\[CurlyPhi]_]:=Map[(GetLabel[g,#])&,\[CurlyPhi]]
(* Get the list of labels of a list of paths given in vertex list representation. *)
PathListLabels[g_,l_]:=Map[(PathLabels[g,#])&,l]



(* Get the labels of a path which is given in edge list representation. *)
EdgePathLabels[g_,\[CurlyPhi]_]:=Map[(GetLabel[g,SourceVertex[#],TargetVertex[#]])&,\[CurlyPhi]]
(* Get the list of labels of a list of paths given in vertex list representation. *)
EdgePathListLabels[g_,l_]:=Map[(EdgePathLabels[g,#])&,l]



(* Utility function for describing paths in terms of edges. *)
SequenceToPairs[{}]:={}
SequenceToPairs[{x_}]:={}
SequenceToPairs[\[Xi]_List]:=Prepend[SequenceToPairs[Rest[\[Xi]]],{\[Xi][[1]],\[Xi][[2]]}]
PathVerticesToEdges[\[CurlyPhi]_List]:=Map[(#[[1]]\[DirectedEdge]#[[2]])&,SequenceToPairs[\[CurlyPhi]]]



(* Utility functions for finding all paths between vertices. *)
FindAllPaths[g_,v_,w_]:=If[VertexList[g]=={},{},FindPath[g,v,w,Infinity,All]]
FindAllEdgePaths[g_,v_,w_]:=Map[PathVerticesToEdges,FindAllPaths[g,v,w]]



(* Find the longest edge path from v to w in g. *)
LongestPath[g_,v_,w_]:=MaximalBy[FindAllEdgePaths[g,v,w],Length]
LongestPathLength[g_,v_,w_]:=Max[Map[Length,FindAllEdgePaths[g,v,w]]]



(* Find all vertices that lie on any path from one vertex to another. *)
VerticesBetween[g_,v_,w_]:=DeleteDuplicates[Catenate[FindAllPaths[g,v,w]]]



(* Find all edges that lie on any path from one vertex to another. *)
EdgesBetween[g_,v_,w_]:=DeleteDuplicates[Catenate[FindAllEdgePaths[g,v,w]]]


(* Analogues to the above specifically for paths from one vertex to another through some specified intermediate vertex. *)
FindAllPathsThrough[g_,v_,i_,w_]:=If[VertexList[g]=={},{},Select[FindPath[g,v,w,Infinity,All],(MemberQ[#,i])&]]
FindAllEdgePathsThrough[g_,v_,i_,w_]:=Map[PathVerticesToEdges,FindAllPathsThrough[g,v,i,w]]
VerticesBetweenThrough[g_,v_,i_,w_]:=DeleteDuplicates[Catenate[FindAllPathsThrough[g,v,i,w]]]
EdgesBetweenThrough[g_,v_,i_,w_]:=DeleteDuplicates[Catenate[FindAllEdgePathsThrough[g,v,i,w]]]



(* Find all the edges within the given list which point to the given vertex. *)
IncomingEdges[v_,es_]:=Select[es,MatchQ[_\[DirectedEdge]v]]
IncomingEdgesList[vs_,es_]:=Map[(IncomingEdges[#,es])&,vs]


(* Find all the source vertices of the edges within the given list which point to the given vertex. *)
IncomingVertices[v_,es_]:=Map[SourceVertex,IncomingEdges[v,es]]
IncomingVerticesList[vs_,es_]:=Map[(IncomingVertices[#,es])&,vs]


(* Find all the edges within the given list which point from the given vertex. *)
OutgoingEdges[v_,es_]:=Select[es,MatchQ[v\[DirectedEdge]_]]
OutgoingEdgesList[vs_,es_]:=Map[(OutgoingEdges[#,es])&,vs]


(* Find all the target vertices of the edges within the given list which point from the given vertex. *)
OutgoingVertices[v_,es_]:=Map[TargetVertex,OutgoingEdges[v,es]]
OutgoingVerticesList[vs_,es_]:=Map[(OutgoingVertices[#,es])&,vs]


(* Find all the edges from the given list that point to or from the given vertex. *)
AdjacentEdges[v_,es_]:=DeleteDuplicates[Join[IncomingEdges[v,es],OutgoingEdges[v,es]]]



(* Remove all the edges adjacent to the given vertex from the given edge list. *)
CutEdges[v_,es_]:=Complement[es,AdjacentEdges[v,es]]


(* Is an edge adjacent to the given vertex? *)
IsAdjacent[v_,e_]:=SourceVertex[e]==v\[Or]TargetVertex[e]==v

(* Determine whether a path contains the given vertex. *)
PathContains[p_,v_]:=AnyTrue[p,(IsAdjacent[v,#])&]



(* Determine whether the given vertex is the start of the given path. *)
PathStart[\[CurlyPhi]_List]:=SourceVertex[\[CurlyPhi][[1]]]
IsPathStart[v_,\[CurlyPhi]_List]:=PathStart[\[CurlyPhi]]==v



(* Determine whether the given vertex is the end of the given path. *)
PathEnd[\[CurlyPhi]_List]:=TargetVertex[\[CurlyPhi][[-1]]]
IsPathEnd[v_,\[CurlyPhi]_List]:=PathEnd[\[CurlyPhi]]==v



(* Does the given graph have any edges? *)
HasEdges[g_]:=EdgeList[g]!={}



(* Produce the subgraph of the given graph with the edge set reduced to the given one. *)
ReduceEdges[g_,es_]:=Module[{edges=Complement[EdgeList[g],es]},
If[edges=={},g,EdgeDelete[g,edges]]]



(* Remove any of the given edges which are in the given graph. *)
RemoveEdges[g_,es_]:=Module[{edges=Intersection[EdgeList[g],es]},
If[edges=={},g,EdgeDelete[g,edges]]]



(* Produce the subgraph of the given graph with the edge set reduced to those on paths from v to w. *)
RestrictEdges[g_,v_,w_]:=ReduceEdges[g,EdgesBetween[g,v,w]]



(* Produce the subgraph of the given graph with the edge set reduced to those on paths from v to w through i. *)
RestrictEdgesThrough[g_,v_,i_,w_]:=ReduceEdges[g,EdgesBetweenThrough[g,v,i,w]]



(* Change the edge list of the given graph to the given one. *)
ChangeEdges[g_,es_]:=EdgeAdd[ReduceEdges[g,es],Complement[es,EdgeList[g]]]



(* The edge lists of all graphs in the given list. *)
EdgeListList[g_]:=Map[EdgeList,g]



(* Graph versions of IncomingEdges and OutgoingEdges. *)
IncomingEdgesG[g_,v_]:=IncomingEdges[v,EdgeList[g]]
OutgoingEdgesG[g_,v_]:=OutgoingEdges[v,EdgeList[g]]



(* Graph versions of IncomingVertices and OutgoingVertices. *)
IncomingVerticesG[g_,v_]:=IncomingVertices[v,EdgeList[g]]
OutgoingVerticesG[g_,v_]:=OutgoingVertices[v,EdgeList[g]]



(* Test whether a node is a source node. *)
IsSourceNode[g_,n_]:=SameQ[IncomingEdgesG[g,n],{}]



(* The set of nodes with no incoming edges. *)
SourceNodes[g_]:=Select[VertexList[g],(IsSourceNode[g,#])&]



(* Find all the cycles in a graph. *)
FindAllCycles[g_]:=FindCycle[g,Infinity,All]



(* Is the given graph acyclic? *)
IsDAG[g_]:=FindAllCycles[g]=={}



(* Find a list of sets of edges the removal of any of which sets would remove all the cycles from the given graph. (In particular, if the given graph is acyclic, this returns an empty list. *)
FindAllCycleRemovalEdgeSets[g_]:=Module[{cs=FindAllCycles[g]},If[cs=={},{{}},DeleteDuplicates[Map[DeleteDuplicates,Tuples[cs]]]]]



(* Find all the maximal acyclic subgraphs of the given graph. *)
FindAcyclicSubgraphs[g_]:=Map[(If[#=={},g,EdgeDelete[g,#]])&,FindAllCycleRemovalEdgeSets[g]]



(* Find the edges within the given graph comprising the maximal source-sink acyclic graphs from the first given node to the second. *)
FindAcyclicSTEdgeSets[g_,v_,w_]:=
Module[{subdags=FindAcyclicSubgraphs[g]},
If[v==w,
(* There's exactly one ST graph from a vertex to itself, which has no edges. *)
{EdgeList[g]},
If[subdags=={},
{{}},
DeleteNonmaximal[Select[Map[(EdgesBetween[#,v,w])&,subdags],UnsameQ[#,{}]&]]]]]



(* Find all the acyclic source-target subgraphs of the given graph with source v and sink w. *)
FindAcyclicSubSTGraphs[g_,v_,w_]:=Module[{gr=RestrictEdges[g,v,w]},Map[(ReduceEdges[gr,#])&,FindAcyclicSTEdgeSets[gr,v,w]]]



(* Is the given graph an ST-DAG with the given source and target? *)
IsSTDAG[g_,v_,w_]:=IsDAG[g]\[And]SetsEqual[EdgeList[g],EdgesBetween[g,v,w]]



(* These are utility functions for TopologicalInduction (see below). They implement a standard topological sort algorithm, enhanced by a caller-specified "visit" routine for each node. *)
TopologicalInductionRemoveOutgoingEdges[visitf_,gorig_,g_,sort_,accum_,s_,n_]:=
Module[{es},
es=OutgoingEdgesG[g,n];
If[es=={},
TopologicalInductionInternal[visitf,gorig,g,sort,accum,s],
Module[{e,nextg,m,nexts},
e=First[es];
nextg=EdgeDelete[g,e];
m=TargetVertex[e];
nexts=If[IsSourceNode[nextg,m],Append[s,m],s];
TopologicalInductionRemoveOutgoingEdges[visitf,gorig,nextg,sort,accum,nexts,n]
]
]
]
TopologicalInductionInternal[visitf_,gorig_,g_,sort_,accum_,{}]:=
(Assert[HasEdges[g]==False];accum)
TopologicalInductionInternal[visitf_,gorig_,g_,sort_,accum_,s_]:=
Module[{n=First[s],t=Rest[s],newsort,parents,sortedparents},
newsort=Append[sort,n];
parents=Map[SourceVertex,IncomingEdgesG[gorig,n]];
Assert[SubsetQ[sort,parents]];
sortedparents=Select[sort,(MemberQ[parents,#])&];
TopologicalInductionRemoveOutgoingEdges[visitf,gorig,g,newsort,visitf[accum,n,sort,sortedparents],t,n]
]



(* Topological induction is an inductive principle which uses a standard algorithm for topological sort but allows the caller to specify a "visit" routine which is called on each node when its position in the topological sort is first established, and to thread the output of that routine through "accum" (the initial value, or base case of the induction, also specified by the caller). The visit routine receives as parameters the latest value of "accum", the node being visited, the topological sort up to (but not including) the node being visited, and the parents of the node being visited in topological sort order. *)
TopologicalInduction[initial_,visitf_,g_]:=
TopologicalInductionInternal[visitf,g,g,{},initial,SourceNodes[g]]



(* Return a topological sort of the given graph, which must be a DAG.  A given DAG may have more than one topological sort, so if the caller's algorithm is intended to produce some single unique result for any given graph, then it must ensure that any topological sort of that graph produces the same result. *)
TopoSort[g_]:=TopologicalInduction[{},Function[{accum,n,predsort,parents},Append[predsort,n]],g]



(* Return an association containing, for each node in the given graph, that node's parents and ancestors (in topological sort order). *)
TopoSortParentsAndAncestors[g_]:=TopologicalInduction[<||>,
Function[{assoc,n,predsort,parents},
Module[{ancestors},
ancestors=DeleteDuplicates[Join[Select[predsort,
Function[{pred},AnyTrue[parents,
Function[{parent},MemberQ[assoc[parent][[2]],pred]]]]],
parents]];
Append[assoc,n->{parents,ancestors}]
]
],
g
]



(* Check whether a vertex is of type "user". *)
VertexIsUser[g_,n_]:=VertexHasType[g,n,"user"]



(* Assert that a vertex is of type "user". *)
AssertIsUser[g_,n_]:=AssertVertexType[g,n,"user"]



(* Add a vertex of type "user" to an existing graph. *)
AddUser[g_,name_]:=VertexAddNextWithProperties[g,{"vType"->"user","name"->name,VertexLabels->ToString[VertexCount[g]+1]<>":"<>name}]



(* A graph containing one user and no propositions. *)
g1:=AddUser[g0,"User 1"]



(* Add a vertex of type "proposition" to an existing graph. *)
AddProp[g_,t_,name_,p_]:=VertexAddNextWithProperties[g,Join[{"vType"->"prop","pType"->t,"name"->name,VertexLabels->ToString[VertexCount[g]+1]<>":"<>name},p]]



(* Check whether a vertex is of type "proposition". *)
VertexIsProp[g_,n_]:=VertexHasType[g,n,"prop"]



(* Assert that a vertex is of type "proposition". *)
AssertIsProp[g_,n_]:=Assert[VertexIsProp[g,n]]



(* Various vertex filters. *)
AllUsers[g_]:=Select[VertexList[g],VertexIsUser[g,#]&]
AllOtherUsers[g_,u_]:=Complement[AllUsers[g],{u}]
AllProps[g_]:=Select[VertexList[g],VertexIsProp[g,#]&]
AllOtherProps[g_,p_]:=Complement[AllProps[g],{p}]



(* Add an atomic proposition to an existing graph. *)
AddAtomic[g_,name_]:=AddProp[g,"atomic",name,{}]



(* Add a "dot" node -- a node that combines belief in an implication with belief in an antecedent to obtain belief in a consequent. *)
AddDot[g_,i_,a_,c_]:=VertexAddNextWithProperties[g,{"vType"->"dot","imp"->i,"ant"->a,"cons"->c,VertexLabels->{"("<>VertexName[g,i]<>")\[CircleDot]("<>VertexName[g,a]<>")"}}]



(* Add a "contradot" node -- a node that combines belief in an implication with disbelief in a consequent to obtain disbelief in an antecedent. *)
AddContradot[g_,i_,a_,c_]:=VertexAddNextWithProperties[g,{"vType"->"contradot","imp"->i,"ant"->a,"cons"->c,VertexLabels->{"("<>VertexName[g,i]<>")\[CircleMinus]("<>VertexName[g,c]<>")"}}]



(* Add trust by a user in another user to an existing graph. *)
AddTrust[g_,u_,w_,v_]:=(AssertIsUser[g,u];AssertIsUser[g,w];
EdgeAddWithProperties[g,u\[DirectedEdge]w,{"eType"->"t","trust"->v,EdgeLabels->("t:"<>ToString[u]<>"\[DirectedEdge]"<>ToString[w])}])



(* Assert that an edge has type "trust". *)
AssertIsTrust[g_,a_,b_]:=AssertEdgeType[g,a,b,"t"]



(* Retrieve the belief vector from a trust edge. If no trust has been explicitly assigned, return Uncertain. *)
GetTrust[g_,e_]:=With[{u=SourceVertex[e],w=TargetVertex[e]},If[EdgeHasProperty[g,"trust",u,w],(AssertIsTrust[g,u,w];EdgeProperty[g,"trust",u,w]),Uncertain]]



(* Add belief by a user in a proposition to an existing graph. *)
AddBelief[g_,u_,n_,v_]:=(AssertIsUser[g,u];AssertIsProp[g,n];
EdgeAddWithProperties[g,u\[DirectedEdge]n,{"eType"->"b","belief"->v,EdgeLabels->("b:"<>ToString[u]<>"\[DirectedEdge]"<>ToString[n])}])



(* Assert that an edge has type "belief". *)
AssertIsBelief[g_,a_,b_]:=AssertEdgeType[g,a,b,"b"]



(* Retrieve the belief vector from a belief edge. If no belief has been explicitly assigned, return Uncertain. *)
GetBelief[g_,e_]:=With[{u=SourceVertex[e],w=TargetVertex[e]},If[EdgeHasProperty[g,"belief",u,w],(AssertIsBelief[g,u,w];EdgeProperty[g,"belief",u,w]),Uncertain]]



GetVector[g_,e_]:=Switch[EdgeType[g,SourceVertex[e],TargetVertex[e]],
"t",GetTrust[g,e],
"b",GetBelief[g,e],
_,Assert[False]]



(* Add the negation (the only unary constructor) of an existing proposition to the graph. *)
AddNewNegationEdge[g_,m_,n_]:=(AssertIsProp[g,m];AssertIsProp[g,n];EdgeAddWithProperties[g,m\[DirectedEdge]n,{"eType"->"n",EdgeLabels->("n:"<>ToString[m]<>"\[DirectedEdge]"<>ToString[n])}])
AddNewNegationEdges[g_,n_]:=AddNewNegationEdge[AddNewNegationEdge[g,VertexCount[g],n],n,VertexCount[g]]
AddNegation[g_,n_]:=(AssertIsProp[g,n];AddNewNegationEdges[AddProp[g,"negation","\[Not]("<>VertexName[g,n]<>")",{"negand"->n}],n])



(* Add a proposition representing (intuitionistic) implication from p_ to q_. In addition to the vertex representing p\[Implies]q itself, this also creates a "dot" node, which propagates belief in p\[Implies]q and belief in p to belief in q, and a "contradot" node, which propagates belief in p\[Implies]q and disbelief in q to disbelief in p, along with all the edges that connect them. *)
AddImplication[g_,p_,q_]:=Module[{h=g,piq=VertexCount[g]+1,pdq=VertexCount[g]+2,pcq=VertexCount[g]+3},(AssertIsProp[g,p];AssertIsProp[g,q];h=AddProp[h,"imp","("<>VertexName[h,p]<>")\[Implies]("<>VertexName[h,q]<>")",{"ant"->p,"cons"->q}];
h=AddDot[h,piq,p,q];
h=AddContradot[h,piq,p,q];
h=EdgeAddWithProperties[h,piq\[DirectedEdge]pdq,{"eType"->"i2d",EdgeLabels->("i2d:"<>ToString[piq]<>"\[DirectedEdge]"<>ToString[pdq])}];h=EdgeAddWithProperties[h,p\[DirectedEdge]pdq,{"eType"->"a2d",EdgeLabels->("a2d:"<>ToString[p]<>"\[DirectedEdge]"<>ToString[pdq])}];h=EdgeAddWithProperties[h,pdq\[DirectedEdge]q,{"eType"->"d2cq",EdgeLabels->("d2cq:"<>ToString[pdq]<>"\[DirectedEdge]"<>ToString[q])}];h=EdgeAddWithProperties[h,piq\[DirectedEdge]pcq,{"eType"->"i2cd",EdgeLabels->("i2cd:"<>ToString[piq]<>"\[DirectedEdge]"<>ToString[pcq])}];h=EdgeAddWithProperties[h,q\[DirectedEdge]pcq,{"eType"->"cq2cd",EdgeLabels->("cq2cd:"<>ToString[q]<>"\[DirectedEdge]"<>ToString[pcq])}];h=EdgeAddWithProperties[h,pcq\[DirectedEdge]p,{"eType"->"cd2a",EdgeLabels->("cd2a:"<>ToString[pcq]<>"\[DirectedEdge]"<>ToString[p])}];
h)]



(* Add implications in both directions between the given propositions. *)
AddImplications[g_,p_,q_]:=AddImplication[AddImplication[g,p,q],q,p]

(* The scalar multiplication operator from the paper. *)
scale[a_,{xb_,xd_,xu_}]:={a*xb,a*xd,xu}/(a*(xb+xd)+xu)


(* The discount operator from the paper, parameterized on the discount-scalar function. *)
discount[dsf_,{vb_,vd_,vu_},w_]:=scale[dsf[{vb,vd,vu}],w]



(* The function that produces the scalar used in the discount operator, called "g" in the paper, which we have more than one possible choice of, with constraints (monotonicity, continuity, a range of [0,1]).  For the purposes of the calculations below, we make a simple choice of the belief component of the vector. *)
defaultDiscountScalar[{b_,d_,u_}]:=b



(* The default discount operator produced by using the default discountScalar operation. *)
defaultDiscount[v_,w_]:=discount[defaultDiscountScalar,v,w]


(* These are "private" utility functions for generating the EBSL recurrence equations. *)
OtherUsersWithEdgesTo[g_,i_,n_]:=Select[AllOtherUsers[g,i],(EdgeQ[g,#\[DirectedEdge]n])&]
PropsWithUserBeliefs[g_]:=Select[AllProps[g],(UnsameQ[Select[IncomingEdgesG[g,#],(VertexIsUser[g,SourceVertex[#]])&],{}])&]
\[ScriptCapitalR]varsToSolveFor[g_,i_]:=Map[(\[ScriptCapitalR][i,#])&,AllOtherUsers[g,i]]
\[ScriptCapitalR]vars[g_,i_]:=\[ScriptCapitalR]varsToSolveFor[g,i]
\[ScriptCapitalR]eq[g_,i_,j_]:=\[ScriptCapitalR][i,j]==\[GothicCapitalC][Join[{GetTrust[g,i\[DirectedEdge]j]},Map[(\[ScriptCapitalR][i,#]\[Cross]GetTrust[g,#\[DirectedEdge]j])&,OtherUsersWithEdgesTo[g,i,j]]]]
\[ScriptCapitalR]eqs[g_,i_]:=Map[(\[ScriptCapitalR]eq[g,i,#])&,AllOtherUsers[g,i]]
\[ScriptCapitalF]varsToSolveFor[g_,i_]:=Map[(\[ScriptCapitalF][i,#])&,PropsWithUserBeliefs[g]]
\[ScriptCapitalF]vars[g_,i_]:=\[ScriptCapitalF]varsToSolveFor[g,i]
\[ScriptCapitalF]eq[g_,i_,p_]:=\[ScriptCapitalF][i,p]==\[GothicCapitalC][Join[{GetBelief[g,i\[DirectedEdge]p]},Map[(\[ScriptCapitalR][i,#]\[Cross]GetBelief[g,#\[DirectedEdge]p])&,OtherUsersWithEdgesTo[g,i,p]]]]
\[ScriptCapitalF]eqs[g_,i_]:=Map[(\[ScriptCapitalF]eq[g,i,#])&,PropsWithUserBeliefs[g]]
AllVariablesToSolveFor[g_,i_]:=Join[\[ScriptCapitalR]varsToSolveFor[g,i],\[ScriptCapitalF]varsToSolveFor[g,i]]
AllVariables[g_,i_]:=Join[\[ScriptCapitalR]vars[g,i],\[ScriptCapitalF]vars[g,i]]
AllMatrixEqs[g_,i_]:=Join[\[ScriptCapitalR]eqs[g,i],\[ScriptCapitalF]eqs[g,i]]
AllValidityEqs[g_,i_]:=Map[(\!\(\*OverscriptBox[\(#\), \(\[RightVector]\)]\))&,AllVariables[g,i]]
AllEqs[g_,i_]:=Join[AllMatrixEqs[g,i],AllValidityEqs[g,i] ]
ExplicitVars[vs_]:=Map[(#->{#[b],#[d],#[u]})&,vs]
EBSLExplicitSubs[g_,i_]:=ExplicitVars[AllVariables[g,i]]
EBSLExplicitVariables[g_,i_]:=Catenate[AllVariables[g,i]/.EBSLExplicitSubs[g,i]]
EBSLExplicitEquations[g_,i_,cs_,subs_]:=EvalBelief[(Join[AllMatrixEqs[g,i],cs]/.subs)/.EBSLExplicitSubs[g,i]]
EBSLExplicitVariablesToSolveFor[g_,i_,subs_]:=Catenate[(AllVariablesToSolveFor[g,i]/.subs)/.EBSLExplicitSubs[g,i]]
SelectSolutions[v_,solns_]:=Map[(#[[Key[v]]])&,Map[Association,solns]]
SelectExplicitSolutions[v_,solns_]:=Map[({#[[Key[v[b]]]],#[[Key[v[d]]]],#[[Key[v[u]]]]})&,Map[Association,solns]]
EBSLNumericEquations[g_,i_,cs_,subs_]:=EvalBelief[(Join[AllEqs[g,i],cs]/.subs)/.EBSLExplicitSubs[g,i]]



(* These are the "public" functions for obtaining solutions to the EBSL equations. The caller may specify additional constraints. To allow the comparison of EBSL and SPIL behaviors in the same contexts, the graph parameters below use the same form as the graphs we have been using in defining and illustrating SPIL. *)
EBSLSymbolicSolutions[g_,i_,cs_]:=Solve[Join[AllMatrixEqs[g,i],cs],AllVariablesToSolveFor[g,i]]
EBSLSymbolicTrust[g_,i_,j_,cs_]:=SelectSolutions[\[ScriptCapitalR][i,j],EBSLSymbolicSolutions[g,i,cs]]
EBSLSymbolicBelief[g_,i_,p_,cs_]:=SelectSolutions[\[ScriptCapitalF][i,p],EBSLSymbolicSolutions[g,i,cs]]
EBSLExplicitSolutions[g_,i_,cs_,subs_]:=Solve[EBSLExplicitEquations[g,i,cs,subs],EBSLExplicitVariablesToSolveFor[g,i,subs]]
EBSLExplicitEquationSelection[g_,i_,v_,cs_,subs_]:=SelectExplicitSolutions[v,EBSLExplicitSolutions[g,i,cs,subs]]
EBSLExplicitTrust[g_,i_,j_,cs_,subs_]:=EBSLExplicitEquationSelection[g,i,\[ScriptCapitalR][i,j],cs,subs]
EBSLExplicitBelief[g_,i_,p_,cs_,subs_]:=EBSLExplicitEquationSelection[g,i,\[ScriptCapitalF][i,p],cs,subs]
EBSLNumericSolutions[g_,i_,cs_,subs_]:=NSolve[EBSLNumericEquations[g,i,cs,subs],EBSLExplicitVariablesToSolveFor[g,i,subs]]
EBSLNumericEquationSelection[g_,i_,v_,cs_,subs_]:=SelectExplicitSolutions[v,EBSLNumericSolutions[g,i,cs,subs]]
EBSLNumericTrust[g_,i_,j_,cs_,subs_]:=EBSLNumericEquationSelection[g,i,\[ScriptCapitalR][i,j],cs,subs]
EBSLNumericBelief[g_,i_,p_,cs_,subs_]:=EBSLNumericEquationSelection[g,i,\[ScriptCapitalF][i,p],cs,subs]
CheckEBSLSymbolicSolutions[g_,i_,cs_,solns_]:=AllTrue[Flatten[(Join[AllMatrixEqs[g,i],cs])/.solns],TrueQ]
CheckEBSLExplicitSolutions[g_,i_,cs_,subs_,solns_]:=AllTrue[Flatten[(EBSLExplicitEquations[g,i,cs,subs])/.solns],TrueQ]
CheckEBSLNumericSolutions[g_,i_,cs_,subs_,solns_]:=AllTrue[Flatten[N[(EBSLNumericEquations[g,i,cs,subs])/.solns]],TrueQ]

(* evidenceToOpinionFunc returns function with provided threshold that converts evidence to opinion vector *)
evidenceToOpinionFunc[c_] := {#1, #2, c}/(#1 + #2 + c) &;

(* rowsToGraph converts rows {{from, to, positive_support, negative_support},} to graph, first argument is function that converts evidence to opinion *)
rowsToGraph[ev2op_, rows_] := Module[{g},
  g = rows //
     Cases[#, {from_, to_, pos_, neg_} ->
        Property[
         DirectedEdge[from,to], {EdgeLabels ->
           StringForm["t:``->``", from, to],
          "eType" -> "t", "trust" -> ev2op[pos, neg]}]] & // Graph;
  PropertyValue[g, VertexLabels] = "Name";
  Scan[(PropertyValue[{g, #}, "vType"] = "user") &, VertexList[g]];
  g
  ]

(* exportEBSLNumericSolution exports solution to file in format {{from, to, discounting}} *)
exportEBSLNumericSolution [filename_, sol_] :=
 sol // Flatten //
   Cases[#, Rule[\[ScriptCapitalR][f_, t_][b], y_] -> {f, t, y}] & //
  Export[filename, #, "Table"] &

Main[threshold_, inputFile_, outputFile_] := Module[{input, solution},

    (* {{from, to, positive_support, negative_support},} *)
    input = Import[inputFile, "Table"];

    solution = rowsToGraph[evidenceToOpinionFunc[threshold], input] // EBSLNumericSolutions[#, 1, {}, {}] &;

    exportEBSLNumericSolution[outputFile, solution];
];

If[Length[$ScriptCommandLine]!=4,
    Print["Provide arguments: <threshold> <input file name> <output file name>"],

    args = Rest[$ScriptCommandLine];
    threshold = ToExpression[args[[1]]];
    inputFileName = args[[2]];
    outputFileName = args[[3]];

    Main[threshold, inputFileName, outputFileName];
];