function TIC()
	local t=time()//32
	for x=-120,119 do
		for y=-68,67 do
   local a=math.atan2(x,y)
			local d=(x^2+y^2)^.5
   local c=1+15*(a/(math.pi*2))
			pix(120+x,68+y,c)
		end
	end 
end
