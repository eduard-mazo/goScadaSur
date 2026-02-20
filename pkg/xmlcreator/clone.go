package xmlcreator

// Clone crea una copia profunda de ElementDef
func (e ElementDef) Clone() ElementDef {
	clone := ElementDef{}
	if e.Analog != nil {
		c := *e.Analog
		if e.Analog.AnalogValue != nil {
			av := *e.Analog.AnalogValue
			c.AnalogValue = &av
		}
		if e.Analog.AnalogInfo != nil {
			ai := *e.Analog.AnalogInfo
			c.AnalogInfo = &ai
		}
		clone.Analog = &c
	}
	if e.Discrete != nil {
		c := *e.Discrete
		if e.Discrete.DiscreteValue != nil {
			dv := *e.Discrete.DiscreteValue
			c.DiscreteValue = &dv
		}
		if e.Discrete.DiscreteInfo != nil {
			di := *e.Discrete.DiscreteInfo
			c.DiscreteInfo = &di
		}
		clone.Discrete = &c
	}
	if e.Breaker != nil {
		c := *e.Breaker
		if e.Breaker.Terminals != nil {
			c.Terminals = make([]*Terminal, len(e.Breaker.Terminals))
			for i, t := range e.Breaker.Terminals {
				if t != nil {
					tc := *t
					c.Terminals[i] = &tc
				}
			}
		}
		if e.Breaker.Discrete != nil {
			dc := *e.Breaker.Discrete
			// Breaker's Discrete might have deep fields too
			if e.Breaker.Discrete.DiscreteValue != nil {
				dv := *e.Breaker.Discrete.DiscreteValue
				dc.DiscreteValue = &dv
			}
			if e.Breaker.Discrete.DiscreteInfo != nil {
				di := *e.Breaker.Discrete.DiscreteInfo
				dc.DiscreteInfo = &di
			}
			c.Discrete = &dc
		}
		clone.Breaker = &c
	}
	if e.IfsPoint != nil {
		c := *e.IfsPoint
		if e.IfsPoint.Link_IfsPointLinksToInfo != nil {
			l := *e.IfsPoint.Link_IfsPointLinksToInfo
			c.Link_IfsPointLinksToInfo = &l
		}
		clone.IfsPoint = &c
	}
	return clone
}
